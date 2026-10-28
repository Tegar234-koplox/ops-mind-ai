package store

import (
	"context"
	"fmt"
	"opsmind/internal/domain"
	"strings"
	"time"
)

type AbsenceStreaks struct {
	Employees []string
	Sources   []domain.SourceRef
}

// Only the last three scheduled working shifts qualify. Leave or an unknown record breaks the streak.
func (s *Store) AbsenceStreaks(ctx context.Context, sh Shift, asOf time.Time) (AbsenceStreaks, error) {
	out := AbsenceStreaks{Employees: []string{}, Sources: []domain.SourceRef{}}
	rows, e := s.Pool.Query(ctx, `WITH history AS (
 SELECT a.employee_id,a.id,sh.id AS shift_id,t.id AS attendance_id,sh.ends_at,
 coalesce(t.source_complete,false) AS complete,t.absence_confirmed_at,
 EXISTS(SELECT 1 FROM leave_records l WHERE l.employee_id=a.employee_id AND l.approved AND l.starts_at<=sh.starts_at AND l.ends_at>=sh.ends_at) AS on_leave,
 row_number() OVER(PARTITION BY a.employee_id ORDER BY sh.ends_at DESC,a.id DESC) AS rn
 FROM shift_assignments a JOIN shift_instances sh ON sh.id=a.shift_id JOIN calendar_days d ON d.day=sh.operational_date
 LEFT JOIN attendance_records t ON t.assignment_id=a.id
 WHERE a.employee_id IN(SELECT employee_id FROM shift_assignments WHERE shift_id=$1) AND d.is_working AND sh.ends_at<=$2)
 SELECT employee_id,id,shift_id,attendance_id,ends_at,complete,absence_confirmed_at,on_leave FROM history WHERE rn<=3 ORDER BY employee_id,rn`, sh.ID, sh.End)
	if e != nil {
		return out, e
	}
	defer rows.Close()
	type observation struct {
		employee, assignment, shift string
		attendance                  *string
		end                         time.Time
		complete                    bool
		confirmed                   *time.Time
		leave                       bool
	}
	groups := map[string][]observation{}
	order := []string{}
	for rows.Next() {
		var v observation
		if e = rows.Scan(&v.employee, &v.assignment, &v.shift, &v.attendance, &v.end, &v.complete, &v.confirmed, &v.leave); e != nil {
			return out, e
		}
		if _, ok := groups[v.employee]; !ok {
			order = append(order, v.employee)
		}
		groups[v.employee] = append(groups[v.employee], v)
	}
	if e = rows.Err(); e != nil {
		return out, e
	}
	for _, employee := range order {
		g := groups[employee]
		if len(g) != 3 || g[0].shift != sh.ID {
			continue
		}
		valid := true
		for _, v := range g {
			if !v.complete || v.leave || v.attendance == nil || v.confirmed == nil || v.confirmed.Before(v.end) || v.confirmed.After(asOf) {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		out.Employees = append(out.Employees, employee)
		for _, v := range g {
			out.Sources = append(out.Sources, domain.SourceRef{Type: "shift_assignments", ID: v.assignment}, domain.SourceRef{Type: "attendance_records", ID: *v.attendance}, domain.SourceRef{Type: "shift_instances", ID: v.shift})
		}
	}
	return out, nil
}

type ApprovalSummary struct {
	Unapproved, Minutes, Unverified, PendingApproval int64
	Sources                                          []domain.SourceRef
}

func (s *Store) OvertimeApprovals(ctx context.Context, sh Shift, asOf time.Time) (ApprovalSummary, error) {
	out := ApprovalSummary{Sources: []domain.SourceRef{}}
	rows, e := s.Pool.Query(ctx, `SELECT o.id,o.minutes,o.complete,coalesce(v.status='approved',o.approved),v.id
 FROM overtime_records o JOIN shift_assignments a ON a.id=o.assignment_id
 LEFT JOIN LATERAL(SELECT id,status FROM overtime_approval_events WHERE overtime_id=o.id AND occurred_at<=$2 ORDER BY occurred_at DESC,id DESC LIMIT 1)v ON true
 WHERE a.shift_id=$1 AND o.minutes>0 ORDER BY o.id`, sh.ID, asOf)
	if e != nil {
		return out, e
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var minutes int64
		var complete, approved bool
		var event *string
		if e = rows.Scan(&id, &minutes, &complete, &approved, &event); e != nil {
			return out, e
		}
		out.Sources = append(out.Sources, domain.SourceRef{Type: "overtime_records", ID: id})
		if event != nil {
			out.Sources = append(out.Sources, domain.SourceRef{Type: "overtime_approval_events", ID: *event})
		}
		if !complete {
			out.Unverified++
			if !approved {
				out.PendingApproval++
			}
			continue
		}
		if !approved {
			out.Unapproved++
			out.Minutes += minutes
		}
	}
	return out, rows.Err()
}

type ExpenseVariance struct {
	Observed, Baseline, Variance int64
	Comparable                   int
	Complete, PostingClosed      bool
	Sources                      []domain.SourceRef
}

func (s *Store) ExpenseVariance(ctx context.Context, sh Shift, asOf time.Time) (ExpenseVariance, error) {
	out := ExpenseVariance{Sources: []domain.SourceRef{{Type: "shift_instances", ID: sh.ID}}}
	var closed bool
	e := s.Pool.QueryRow(ctx, `SELECT coalesce(sum(x.amount_rupiah) FILTER(WHERE x.occurred_at<=$2),0),coalesce(bool_and(x.complete AND x.occurred_at<=$2),true),EXISTS(SELECT 1 FROM expense_closures WHERE shift_id=$1 AND closed_at<=$2 AND closed_at>=$3) FROM expenses x WHERE x.shift_id=$1`, sh.ID, asOf, sh.End).Scan(&out.Observed, &out.Complete, &closed)
	if e != nil {
		return out, e
	}
	out.PostingClosed = closed
	out.Complete = out.Complete && closed && sh.Complete && sh.Planned > 0
	if closed {
		out.Sources = append(out.Sources, domain.SourceRef{Type: "expense_closures", ID: sh.ID})
	}
	rows, e := s.Pool.Query(ctx, `SELECT sh.id,p.planned,coalesce(sum(x.amount_rupiah),0) FROM shift_instances sh
 JOIN production_runs p ON p.shift_id=sh.id JOIN calendar_days d ON d.day=sh.operational_date
 JOIN expense_closures c ON c.shift_id=sh.id LEFT JOIN expenses x ON x.shift_id=sh.id
 WHERE sh.line_id=$1 AND sh.code=$2 AND sh.product=$3 AND sh.ends_at<=$4 AND sh.complete AND p.complete AND p.planned>0 AND d.is_working AND c.closed_at<=$5 AND c.closed_at>=sh.ends_at
 AND sh.ends_at-sh.starts_at=$6::interval AND (extract(isodow FROM sh.operational_date)=6)=$7
 GROUP BY sh.id,p.planned HAVING coalesce(bool_and(x.complete AND x.occurred_at<=$5),true) ORDER BY sh.starts_at DESC LIMIT 4`, sh.LineID, sh.Code, sh.Product, sh.Start, asOf, fmt.Sprintf("%d seconds", int64(sh.End.Sub(sh.Start)/time.Second)), sh.Start.In(time.FixedZone("Asia/Bangkok", 25200)).Weekday() == time.Saturday)
	if e != nil {
		return out, e
	}
	costs, targets := []int64{}, []int64{}
	ids := []string{sh.ID}
	for rows.Next() {
		var id string
		var cost, target int64
		if e = rows.Scan(&id, &target, &cost); e != nil {
			rows.Close()
			return out, e
		}
		costs = append(costs, cost)
		targets = append(targets, target)
		ids = append(ids, id)
		out.Sources = append(out.Sources, domain.SourceRef{Type: "shift_instances", ID: id}, domain.SourceRef{Type: "expense_closures", ID: id})
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return out, e
	}
	out.Comparable = len(costs)
	refs, e := s.Pool.Query(ctx, "SELECT id FROM expenses WHERE shift_id=ANY($1) AND occurred_at<=$2 ORDER BY id", ids, asOf)
	if e != nil {
		return out, e
	}
	for refs.Next() {
		var id string
		if e = refs.Scan(&id); e != nil {
			refs.Close()
			return out, e
		}
		out.Sources = append(out.Sources, domain.SourceRef{Type: "expenses", ID: id})
	}
	e = refs.Err()
	refs.Close()
	if e != nil {
		return out, e
	}
	out.Baseline, e = domain.AdjustedMedian(costs, targets, sh.Planned)
	if e != nil {
		out.Complete = false
		return out, nil
	}
	out.Variance, e = domain.BPS(out.Observed, out.Baseline)
	if e != nil {
		out.Complete = false
	}
	return out, nil
}

// P1 history is allowed only when the deterministic query selected it for this case.
func (s *Store) ValidateCaseSources(ctx context.Context, c domain.Case, refs []domain.SourceRef) error {
	extra := map[domain.SourceRef]bool{}
	sh, e := s.Shift(ctx, c.ShiftID)
	if e != nil {
		return e
	}
	var allowed []domain.SourceRef
	switch c.RuleID {
	case "ATT-001":
		v, err := s.AbsenceStreaks(ctx, sh, domain.AsOf)
		e = err
		allowed = v.Sources
	case "OT-002":
		v, err := s.OvertimeApprovals(ctx, sh, domain.AsOf)
		e = err
		allowed = v.Sources
	case "COST-001", "COST-002":
		v, err := s.ExpenseVariance(ctx, sh, domain.AsOf)
		e = err
		allowed = v.Sources
	}
	if e != nil {
		return e
	}
	for _, r := range allowed {
		extra[r] = true
	}
	for _, r := range refs {
		if !extra[r] {
			if e = s.ValidateSources(ctx, c.ShiftID, []domain.SourceRef{r}); e != nil {
				return e
			}
		}
	}
	return nil
}
func EmployeeIDs(v AbsenceStreaks) string { return strings.Join(v.Employees, ", ") }
