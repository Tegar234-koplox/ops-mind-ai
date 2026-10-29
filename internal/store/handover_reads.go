package store

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"opsmind/internal/domain"
	"sort"
	"time"
)

var ErrShiftNotFinished = errors.New("shift belum selesai pada waktu evaluasi")

const handoverShiftCols = `sh.id,sh.line_id,l.name,to_char(sh.operational_date,'YYYY-MM-DD'),sh.code,sh.product,sh.starts_at,sh.ends_at,d.is_working`
const handoverShiftJoin = ` FROM shift_instances sh JOIN production_lines l ON l.id=sh.line_id JOIN calendar_days d ON d.day=sh.operational_date `

func scanHandoverShift(row pgx.Row) (domain.HandoverShift, error) {
	var v domain.HandoverShift
	err := row.Scan(&v.ID, &v.LineID, &v.LineName, &v.OperationalDate, &v.Code, &v.Product, &v.StartsAt, &v.EndsAt, &v.Working)
	return v, normalize(err)
}

type LineOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Store) HandoverLines(ctx context.Context) ([]LineOption, error) {
	rows, err := s.readDB().Query(ctx, "SELECT id,name FROM production_lines WHERE facility_id=$1 ORDER BY id", domain.FacilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LineOption{}
	for rows.Next() {
		var v LineOption
		if err = rows.Scan(&v.ID, &v.Name); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) HandoverShifts(ctx context.Context, line, date string, limit, offset int, asOf time.Time) ([]domain.HandoverShift, int, error) {
	where := ` WHERE l.facility_id=$1 AND sh.ends_at<=$2 AND ($3='' OR sh.line_id=$3) AND ($4='' OR sh.operational_date=NULLIF($4,'')::date)`
	args := []any{domain.FacilityID, asOf, line, date}
	var total int
	if err := s.readDB().QueryRow(ctx, "SELECT count(*)"+handoverShiftJoin+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.readDB().Query(ctx, "SELECT "+handoverShiftCols+handoverShiftJoin+where+" ORDER BY sh.starts_at DESC,sh.line_id,sh.id LIMIT $5 OFFSET $6", append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.HandoverShift{}
	for rows.Next() {
		v, err := scanHandoverShift(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

type HandoverData struct {
	AsOf           time.Time
	Shift          domain.HandoverShift
	NextShift      *domain.HandoverShift
	Operation      Shift
	Staff          Staffing
	Overtime       Overtime
	Approval       ApprovalSummary
	Expenses       ExpenseVariance
	Downtime       int64
	DowntimeEvents []domain.HandoverDowntime
	Cases          []domain.HandoverCase
	Sources        []domain.SourceRef
}

func (s *Store) HandoverData(ctx context.Context, shiftID string, asOf time.Time) (HandoverData, error) {
	data := HandoverData{AsOf: asOf, Sources: []domain.SourceRef{}}
	err := s.readSnapshot(ctx, func(reader *Store) error {
		var err error
		data.Shift, err = scanHandoverShift(reader.readDB().QueryRow(ctx, "SELECT "+handoverShiftCols+handoverShiftJoin+" WHERE sh.id=$1 AND l.facility_id=$2", shiftID, domain.FacilityID))
		if err != nil {
			return err
		}
		if data.Shift.EndsAt.After(asOf) {
			return ErrShiftNotFinished
		}
		next, err := scanHandoverShift(reader.readDB().QueryRow(ctx, "SELECT "+handoverShiftCols+handoverShiftJoin+" WHERE sh.line_id=$1 AND sh.starts_at=$2 AND l.facility_id=$3 ORDER BY sh.id LIMIT 1", data.Shift.LineID, data.Shift.EndsAt, domain.FacilityID))
		if err == nil {
			data.NextShift = &next
		} else if !errors.Is(err, ErrNotFound) {
			return err
		}
		if data.Operation, err = reader.Shift(ctx, shiftID); err != nil {
			return err
		}
		if data.Staff, err = reader.staffingAt(ctx, shiftID, asOf); err != nil {
			return err
		}
		if data.Overtime, err = reader.Overtime(ctx, data.Operation); err != nil {
			return err
		}
		if data.Approval, err = reader.OvertimeApprovals(ctx, data.Operation, asOf); err != nil {
			return err
		}
		if data.Expenses, err = reader.ExpenseVariance(ctx, data.Operation, asOf); err != nil {
			return err
		}
		var refs []domain.SourceRef
		if data.Downtime, refs, err = reader.Downtime(ctx, data.Operation); err != nil {
			return err
		}
		data.Sources = append(data.Sources, refs...)
		if data.DowntimeEvents, err = reader.handoverDowntimeEvents(ctx, shiftID); err != nil {
			return err
		}
		if data.Cases, err = reader.handoverCases(ctx, shiftID); err != nil {
			return err
		}
		data.Sources = append(data.Sources, data.Staff.Sources...)
		data.Sources = append(data.Sources, data.Overtime.Sources...)
		data.Sources = append(data.Sources, data.Approval.Sources...)
		data.Sources = append(data.Sources, data.Expenses.Sources...)
		return reader.handoverSources(ctx, &data)
	})
	return data, err
}

func (s *Store) handoverDowntimeEvents(ctx context.Context, id string) ([]domain.HandoverDowntime, error) {
	rows, err := s.readDB().Query(ctx, `SELECT d.id,m.id,m.name,d.starts_at,d.ends_at,d.reason FROM downtime_events d JOIN machines m ON m.id=d.machine_id JOIN shift_instances sh ON sh.id=d.shift_id WHERE sh.id=$1 AND m.line_id=sh.line_id ORDER BY d.starts_at,d.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.HandoverDowntime{}
	for rows.Next() {
		var v domain.HandoverDowntime
		if err = rows.Scan(&v.ID, &v.MachineID, &v.MachineName, &v.StartsAt, &v.EndsAt, &v.Reason); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) handoverCases(ctx context.Context, id string) ([]domain.HandoverCase, error) {
	rows, err := s.readDB().Query(ctx, "SELECT "+caseCols+" FROM anomaly_cases WHERE shift_id=$1 ORDER BY CASE severity WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END,rule_id,id", id)
	if err != nil {
		return nil, err
	}
	cases := []domain.Case{}
	for rows.Next() {
		c, err := scanCase(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		cases = append(cases, c)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	out := []domain.HandoverCase{}
	for _, c := range cases {
		item := domain.HandoverCase{CaseID: c.ID, RuleID: c.RuleID, Severity: c.Severity, Status: c.Status, DataQuality: c.DataQuality, Explanation: c.Explanation, Recommendations: []domain.Recommendation{}, Evidence: []domain.HandoverEvidenceRef{}}
		run, err := s.LatestRun(ctx, c.ID)
		if errors.Is(err, ErrNotFound) {
			out = append(out, item)
			continue
		}
		if err != nil {
			return nil, err
		}
		item.RunID, item.RunStatus, item.ModelMode = run.ID, run.Status, run.Mode
		if run.Report != nil {
			item.Synthesis = run.Report.Summary
		}
		recs, err := s.Recommendations(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		for _, rec := range recs {
			if rec.RunID == run.ID {
				item.Recommendations = append(item.Recommendations, rec)
			}
		}
		evidence, err := s.EvidenceList(ctx, run.ID)
		if err != nil {
			return nil, err
		}
		for _, ev := range evidence {
			item.Evidence = append(item.Evidence, domain.HandoverEvidenceRef{ID: ev.ID, Hash: ev.Hash})
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *Store) handoverSources(ctx context.Context, data *HandoverData) error {
	ids := []string{data.Shift.ID}
	for _, ref := range append(append([]domain.SourceRef{}, data.Overtime.Sources...), data.Expenses.Sources...) {
		if ref.Type == "shift_instances" {
			ids = append(ids, ref.ID)
		}
	}
	rows, err := s.readDB().Query(ctx, `SELECT 'production_runs',id FROM production_runs WHERE shift_id=ANY($1)
 UNION ALL SELECT 'overtime_records',o.id FROM overtime_records o JOIN shift_assignments a ON a.id=o.assignment_id WHERE a.shift_id=ANY($1)
 UNION ALL SELECT 'leave_records',l.id FROM leave_records l JOIN shift_assignments a ON a.employee_id=l.employee_id JOIN shift_instances sh ON sh.id=a.shift_id WHERE sh.id=$2 AND l.approved AND l.starts_at<=sh.starts_at AND l.ends_at>=sh.ends_at`, ids, data.Shift.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var ref domain.SourceRef
		if err = rows.Scan(&ref.Type, &ref.ID); err != nil {
			rows.Close()
			return err
		}
		data.Sources = append(data.Sources, ref)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, c := range data.Cases {
		data.Sources = append(data.Sources, domain.SourceRef{Type: "anomaly_cases", ID: c.CaseID})
		if c.RunID != "" {
			data.Sources = append(data.Sources, domain.SourceRef{Type: "investigation_runs", ID: c.RunID})
		}
		for _, ev := range c.Evidence {
			data.Sources = append(data.Sources, domain.SourceRef{Type: "case_evidence", ID: ev.ID})
		}
		for _, rec := range c.Recommendations {
			data.Sources = append(data.Sources, domain.SourceRef{Type: "recommendations", ID: rec.ID})
		}
	}
	for _, ev := range data.DowntimeEvents {
		data.Sources = append(data.Sources, domain.SourceRef{Type: "machines", ID: ev.MachineID})
	}
	data.Sources = append(data.Sources, domain.SourceRef{Type: "production_lines", ID: data.Shift.LineID}, domain.SourceRef{Type: "calendar_days", ID: data.Shift.OperationalDate})
	if data.NextShift != nil {
		data.Sources = append(data.Sources, domain.SourceRef{Type: "shift_instances", ID: data.NextShift.ID})
	}
	recIDs := []string{}
	for _, c := range data.Cases {
		for _, rec := range c.Recommendations {
			recIDs = append(recIDs, rec.ID)
		}
	}
	reviews, err := s.readDB().Query(ctx, "SELECT id FROM review_actions WHERE recommendation_id=ANY($1) ORDER BY id", recIDs)
	if err != nil {
		return err
	}
	for reviews.Next() {
		var id string
		if err = reviews.Scan(&id); err != nil {
			reviews.Close()
			return err
		}
		data.Sources = append(data.Sources, domain.SourceRef{Type: "review_actions", ID: id})
	}
	err = reviews.Err()
	reviews.Close()
	if err != nil {
		return err
	}
	seen := map[domain.SourceRef]bool{}
	refs := []domain.SourceRef{}
	for _, ref := range data.Sources {
		if !seen[ref] {
			seen[ref] = true
			refs = append(refs, ref)
		}
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Type != refs[j].Type {
			return refs[i].Type < refs[j].Type
		}
		return refs[i].ID < refs[j].ID
	})
	data.Sources = refs
	return nil
}
