package seed

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"time"
)

const P1AttendanceShift = "B-20261026-S1"
const P1OvertimeShift = "B-20261026-S2"
const P1CostHighShift = "B-20261026-S3"
const P1CostLowShift = "C-20261026-S3"

// RunP1 extends only the known synthetic dataset. It is explicit and transactional, never part of a migration.
func RunP1(ctx context.Context, s *store.Store) (bool, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return false, e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(2026102701)"); e != nil {
		return false, e
	}
	var applied bool
	if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM audit_events WHERE action='synthetic_p1_seeded' AND details->>'version'='20261028-v1')").Scan(&applied); e != nil {
		return false, e
	}
	if applied {
		return false, nil
	}
	var count int
	var golden bool
	if e = tx.QueryRow(ctx, "SELECT (SELECT count(*) FROM employees),EXISTS(SELECT 1 FROM shift_instances WHERE id=$1)", GoldenShift).Scan(&count, &golden); e != nil {
		return false, e
	}
	if count != 150 || !golden {
		return false, fmt.Errorf("seed-p1 hanya untuk dataset sintetis OpsMind seed 20261027")
	}
	for _, apply := range []func(context.Context, pgx.Tx) error{seedConfirmedAbsences, seedOvertimeRevocation, seedExpenseCases} {
		if e = apply(ctx, tx); e != nil {
			return false, e
		}
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_events(actor,action,details) VALUES('seed','synthetic_p1_seeded','{"version":"20261028-v1","synthetic":true}')`); e != nil {
		return false, e
	}
	return true, tx.Commit(ctx)
}

func seedConfirmedAbsences(ctx context.Context, tx pgx.Tx) error {
	var e error
	var employee string
	if e = tx.QueryRow(ctx, "SELECT employee_id FROM shift_assignments WHERE shift_id=$1 ORDER BY id LIMIT 1", P1AttendanceShift).Scan(&employee); e != nil {
		return e
	}
	rows, e := tx.Query(ctx, `SELECT a.id,sh.ends_at FROM shift_assignments a JOIN shift_instances sh ON sh.id=a.shift_id JOIN calendar_days d ON d.day=sh.operational_date WHERE a.employee_id=$1 AND d.is_working AND sh.ends_at<=(SELECT ends_at FROM shift_instances WHERE id=$2) ORDER BY sh.ends_at DESC,a.id DESC LIMIT 3`, employee, P1AttendanceShift)
	if e != nil {
		return e
	}
	type absence struct {
		id  string
		end time.Time
	}
	absences := []absence{}
	for rows.Next() {
		var a absence
		if e = rows.Scan(&a.id, &a.end); e != nil {
			rows.Close()
			return e
		}
		absences = append(absences, a)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	if len(absences) != 3 {
		return fmt.Errorf("fixture ATT-001 perlu tiga penugasan")
	}
	for _, a := range absences {
		if _, e = tx.Exec(ctx, "UPDATE attendance_records SET check_in=NULL,check_out=NULL,verified=false,source_complete=true,absence_confirmed_at=$2,absence_confirmed_by='SUP-SYNTHETIC' WHERE assignment_id=$1", a.id, a.end.Add(5*time.Minute)); e != nil {
			return e
		}
	}

	return nil
}

func seedOvertimeRevocation(ctx context.Context, tx pgx.Tx) error {
	var e error
	var overtime string
	var begin, end time.Time
	if e = tx.QueryRow(ctx, "SELECT o.id,sh.starts_at,sh.ends_at FROM overtime_records o JOIN shift_assignments a ON a.id=o.assignment_id JOIN shift_instances sh ON sh.id=a.shift_id WHERE sh.id=$1 ORDER BY o.id LIMIT 1", P1OvertimeShift).Scan(&overtime, &begin, &end); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO overtime_approval_events(id,overtime_id,status,occurred_at,actor) VALUES('P1-OT-APPROVED',$1,'approved',$2,'SUP-SYNTHETIC'),('P1-OT-REVOKED',$1,'revoked',$3,'SUP-SYNTHETIC')", overtime, begin.Add(-time.Hour), end.Add(-15*time.Minute)); e != nil {
		return e
	}

	return nil
}

func seedExpenseCases(ctx context.Context, tx pgx.Tx) error {
	var e error
	for _, target := range []string{P1CostHighShift, P1CostLowShift} {
		var line, code, product string
		var start, finish time.Time
		var planned int64
		if e = tx.QueryRow(ctx, "SELECT s.line_id,s.code,s.product,s.starts_at,s.ends_at,p.planned FROM shift_instances s JOIN production_runs p ON p.shift_id=s.id WHERE s.id=$1", target).Scan(&line, &code, &product, &start, &finish, &planned); e != nil {
			return e
		}
		r, e := tx.Query(ctx, `SELECT sh.id,p.planned,sum(x.amount_rupiah),sh.ends_at FROM shift_instances sh JOIN production_runs p ON p.shift_id=sh.id JOIN calendar_days d ON d.day=sh.operational_date JOIN expenses x ON x.shift_id=sh.id
   WHERE sh.line_id=$1 AND sh.code=$2 AND sh.product=$3 AND sh.ends_at<=$4 AND sh.complete AND p.complete AND d.is_working AND p.planned>0 AND sh.ends_at-sh.starts_at=interval '8 hours' AND extract(isodow FROM sh.operational_date)<>6
   GROUP BY sh.id,p.planned ORDER BY sh.starts_at DESC LIMIT 4`, line, code, product, start)
		if e != nil {
			return e
		}
		type comparable struct {
			id  string
			end time.Time
		}
		comparables := []comparable{}
		costs, targets := []int64{}, []int64{}
		for r.Next() {
			var c comparable
			var cost, plan int64
			if e = r.Scan(&c.id, &plan, &cost, &c.end); e != nil {
				r.Close()
				return e
			}
			comparables = append(comparables, c)
			costs = append(costs, cost)
			targets = append(targets, plan)
		}
		e = r.Err()
		r.Close()
		if e != nil {
			return e
		}
		baseline, e := domain.AdjustedMedian(costs, targets, planned)
		if e != nil {
			return e
		}
		for _, c := range append(comparables, comparable{id: target, end: finish}) {
			if _, e = tx.Exec(ctx, "INSERT INTO expense_closures(shift_id,closed_at,closed_by) VALUES($1,$2,'FIN-SYNTHETIC')", c.id, c.end.Add(5*time.Minute)); e != nil {
				return e
			}
		}
		desired := baseline * 13 / 10
		if target == P1CostLowShift {
			desired = baseline * 3 / 4
		}
		var other int64
		if e = tx.QueryRow(ctx, "SELECT coalesce(sum(amount_rupiah),0) FROM expenses WHERE shift_id=$1 AND category<>'maintenance'", target).Scan(&other); e != nil {
			return e
		}
		if desired < other {
			return fmt.Errorf("fixture pengeluaran tidak konsisten")
		}
		if _, e = tx.Exec(ctx, "UPDATE expenses SET amount_rupiah=$2 WHERE shift_id=$1 AND category='maintenance'", target, desired-other); e != nil {
			return e
		}
	}

	return nil
}
