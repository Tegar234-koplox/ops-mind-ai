package store

import (
	"context"
	"fmt"
	"opsmind/internal/domain"
	"strconv"
	"time"
)

type Shift struct {
	ID, LineID, Code, Product string
	Start, End                time.Time
	Required, Planned, Actual int64
	Complete, Working         bool
}

func (s *Store) Shift(ctx context.Context, id string) (Shift, error) {
	var v Shift
	e := s.readDB().QueryRow(ctx, "SELECT s.id,s.line_id,s.code,s.product,s.starts_at,s.ends_at,s.required_staff,p.planned,p.actual,s.complete AND p.complete,c.is_working FROM shift_instances s JOIN production_runs p ON p.shift_id=s.id JOIN calendar_days c ON c.day=s.operational_date WHERE s.id=$1", id).Scan(&v.ID, &v.LineID, &v.Code, &v.Product, &v.Start, &v.End, &v.Required, &v.Planned, &v.Actual, &v.Complete, &v.Working)
	return v, normalize(e)
}

type Staffing struct {
	Scheduled, Present, Leave, Unverified, ConfirmedAbsent int64
	Sources                                                []domain.SourceRef
}

func (s *Store) Staffing(ctx context.Context, id string) (Staffing, error) {
	return s.staffingAt(ctx, id, domain.AsOf)
}

func (s *Store) staffingAt(ctx context.Context, id string, asOf time.Time) (Staffing, error) {
	var out Staffing
	out.Sources = []domain.SourceRef{}
	rows, e := s.readDB().Query(ctx, `SELECT a.id,t.id,t.check_in,t.check_out,coalesce(t.verified,false),
 EXISTS(SELECT 1 FROM leave_records l WHERE l.employee_id=a.employee_id AND l.approved AND l.starts_at<=sh.starts_at AND l.ends_at>=sh.ends_at),
 coalesce(t.source_complete AND t.absence_confirmed_at>=sh.ends_at AND t.absence_confirmed_at<=$2,false)
 FROM shift_assignments a JOIN shift_instances sh ON sh.id=a.shift_id LEFT JOIN attendance_records t ON t.assignment_id=a.id WHERE a.shift_id=$1 ORDER BY a.id`, id, asOf)
	if e != nil {
		return out, e
	}
	defer rows.Close()
	for rows.Next() {
		var aid string
		var tid *string
		var in, outTime *time.Time
		var verified, leave, absent bool
		if e = rows.Scan(&aid, &tid, &in, &outTime, &verified, &leave, &absent); e != nil {
			return out, e
		}
		out.Scheduled++
		out.Sources = append(out.Sources, domain.SourceRef{Type: "shift_assignments", ID: aid})
		if tid != nil {
			out.Sources = append(out.Sources, domain.SourceRef{Type: "attendance_records", ID: *tid})
		}
		if leave {
			out.Leave++
		} else if absent {
			out.ConfirmedAbsent++
		} else if verified && in != nil && outTime != nil && outTime.After(*in) {
			out.Present++
		} else {
			out.Unverified++
		}
	}
	return out, rows.Err()
}

type Overtime struct {
	Observed, Baseline, Variance int64
	Comparable                   int
	Complete                     bool
	Sources                      []domain.SourceRef
}

func (s *Store) Overtime(ctx context.Context, sh Shift) (Overtime, error) {
	v := Overtime{Sources: []domain.SourceRef{{Type: "shift_instances", ID: sh.ID}}, Complete: true}
	var missing bool
	e := s.readDB().QueryRow(ctx, "SELECT coalesce(sum(o.cost_rupiah),0),coalesce(bool_or(NOT o.complete),false),count(o.id) FROM shift_assignments a LEFT JOIN overtime_records o ON o.assignment_id=a.id WHERE a.shift_id=$1", sh.ID).Scan(&v.Observed, &missing, &v.Comparable)
	if e != nil {
		return v, e
	}
	v.Complete = !missing && v.Comparable > 0
	rows, e := s.readDB().Query(ctx, `SELECT s.id,p.planned,sum(o.cost_rupiah) FROM shift_instances s JOIN calendar_days c ON c.day=s.operational_date JOIN production_runs p ON p.shift_id=s.id JOIN shift_assignments a ON a.shift_id=s.id JOIN overtime_records o ON o.assignment_id=a.id
 WHERE s.line_id=$1 AND s.code=$2 AND s.product=$3 AND s.ends_at<=$4 AND s.complete AND p.complete AND c.is_working AND p.planned>0
 AND s.ends_at-s.starts_at=$5::interval
 AND (extract(isodow FROM s.operational_date)=6)=$6
 GROUP BY s.id,p.planned HAVING bool_and(o.complete) ORDER BY s.starts_at DESC LIMIT 4`, sh.LineID, sh.Code, sh.Product, sh.Start, fmt.Sprintf("%d seconds", int64(sh.End.Sub(sh.Start)/time.Second)), sh.Start.In(time.FixedZone("Asia/Bangkok", 25200)).Weekday() == time.Saturday)
	if e != nil {
		return v, e
	}
	defer rows.Close()
	costs, targets := []int64{}, []int64{}
	for rows.Next() {
		var id string
		var cost, target int64
		if e = rows.Scan(&id, &target, &cost); e != nil {
			return v, e
		}
		costs = append(costs, cost)
		targets = append(targets, target)
		v.Sources = append(v.Sources, domain.SourceRef{Type: "shift_instances", ID: id})
	}
	if e = rows.Err(); e != nil {
		return v, e
	}
	v.Comparable = len(costs)
	v.Baseline, e = domain.AdjustedMedian(costs, targets, sh.Planned)
	if e != nil {
		v.Complete = false
		return v, nil
	}
	v.Variance, e = domain.BPS(v.Observed, v.Baseline)
	if e != nil {
		v.Complete = false
	}
	return v, nil
}
func (s *Store) Downtime(ctx context.Context, sh Shift) (int64, []domain.SourceRef, error) {
	rows, e := s.readDB().Query(ctx, "SELECT d.id,d.starts_at,d.ends_at FROM downtime_events d JOIN machines m ON m.id=d.machine_id WHERE d.shift_id=$1 AND m.line_id=$2 ORDER BY d.starts_at", sh.ID, sh.LineID)
	if e != nil {
		return 0, nil, e
	}
	defer rows.Close()
	intervals := []domain.Interval{}
	src := []domain.SourceRef{{Type: "shift_instances", ID: sh.ID}}
	for rows.Next() {
		var id string
		var v domain.Interval
		if e = rows.Scan(&id, &v.Start, &v.End); e != nil {
			return 0, nil, e
		}
		intervals = append(intervals, v)
		src = append(src, domain.SourceRef{Type: "downtime_events", ID: id})
	}
	return domain.UnionMinutes(intervals, sh.Start, sh.End), src, rows.Err()
}
func (s *Store) ValidateSources(ctx context.Context, shiftID string, src []domain.SourceRef) error {
	queries := map[string]string{
		"shift_instances":    "SELECT EXISTS(SELECT 1 FROM shift_instances x JOIN shift_instances scope ON scope.id=$2 WHERE x.id=$1 AND x.line_id=scope.line_id AND x.code=scope.code AND x.starts_at<=scope.starts_at)",
		"shift_assignments":  "SELECT EXISTS(SELECT 1 FROM shift_assignments WHERE id=$1 AND shift_id=$2)",
		"attendance_records": "SELECT EXISTS(SELECT 1 FROM attendance_records t JOIN shift_assignments a ON a.id=t.assignment_id WHERE t.id=$1 AND a.shift_id=$2)",
		"leave_records":      "SELECT EXISTS(SELECT 1 FROM leave_records l JOIN shift_assignments a ON a.employee_id=l.employee_id WHERE l.id=$1 AND a.shift_id=$2)",
		"overtime_records":   "SELECT EXISTS(SELECT 1 FROM overtime_records o JOIN shift_assignments a ON a.id=o.assignment_id WHERE o.id=$1 AND a.shift_id=$2)",
		"production_runs":    "SELECT EXISTS(SELECT 1 FROM production_runs WHERE id=$1 AND shift_id=$2)",
		"downtime_events":    "SELECT EXISTS(SELECT 1 FROM downtime_events WHERE id=$1 AND shift_id=$2)",
		"expenses":           "SELECT EXISTS(SELECT 1 FROM expenses WHERE id=$1 AND shift_id=$2)",
	}
	for _, r := range src {
		q, ok := queries[r.Type]
		if !ok {
			return fmt.Errorf("source type tidak diizinkan")
		}
		var found bool
		if e := s.readDB().QueryRow(ctx, q, r.ID, shiftID).Scan(&found); e != nil {
			return e
		}
		if !found {
			return fmt.Errorf("source tidak berada dalam scope")
		}
	}
	return nil
}
func Num(n int64) string { return strconv.FormatInt(n, 10) }
