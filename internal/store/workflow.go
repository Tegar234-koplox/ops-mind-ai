package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"opsmind/internal/domain"
)

func (s *Store) InsertCase(ctx context.Context, c domain.Case, mode string) (domain.Case, bool, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return c, false, e
	}
	defer tx.Rollback(ctx)
	tag, e := tx.Exec(ctx, "INSERT INTO anomaly_cases(id,rule_id,entity_id,shift_id,department,severity,status,data_quality,expected,observed,explanation,sources) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT(rule_id,entity_id,shift_id,rule_version) DO NOTHING", c.ID, c.RuleID, c.EntityID, c.ShiftID, c.Department, c.Severity, c.Status, c.DataQuality, JSON(c.Expected), JSON(c.Observed), c.Explanation, JSON(c.Sources))
	if e != nil {
		return c, false, e
	}
	created := tag.RowsAffected() > 0
	if created {
		if _, e = tx.Exec(ctx, "INSERT INTO audit_events(case_id,actor,action,details) VALUES($1,'detector','case_detected',$2)", c.ID, JSON(map[string]any{"rule_id": c.RuleID, "data_quality": c.DataQuality})); e != nil {
			return c, false, e
		}
		if c.RuleID == "OT-001" && c.DataQuality == "confirmed" {
			if _, e = enqueue(ctx, tx, c.ID, mode, "", "detector"); e != nil {
				return c, false, e
			}
		}
	} else {
		if e = tx.QueryRow(ctx, "SELECT id FROM anomaly_cases WHERE rule_id=$1 AND entity_id=$2 AND shift_id=$3 AND rule_version=1", c.RuleID, c.EntityID, c.ShiftID).Scan(&c.ID); e != nil {
			return c, false, e
		}
	}
	if e = tx.Commit(ctx); e != nil {
		return c, false, e
	}
	c, e = s.Case(ctx, c.ID)
	return c, created, e
}
func enqueue(ctx context.Context, tx pgx.Tx, caseID, mode, instruction, actor string) (string, error) {
	var locked string
	if e := tx.QueryRow(ctx, "SELECT id FROM anomaly_cases WHERE id=$1 FOR UPDATE", caseID).Scan(&locked); e != nil {
		return "", normalize(e)
	}
	var active bool
	if e := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM investigation_runs WHERE case_id=$1 AND status IN('queued','running'))", caseID).Scan(&active); e != nil {
		return "", e
	}
	if active {
		return "", ErrConflict
	}
	id := domain.ID()
	if _, e := tx.Exec(ctx, "INSERT INTO investigation_runs(id,case_id,status,mode,instruction) VALUES($1,$2,'queued',$3,$4)", id, caseID, mode, instruction); e != nil {
		return "", e
	}
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{"UPDATE anomaly_cases SET status='investigating' WHERE id=$1", []any{caseID}},
		{"INSERT INTO investigation_events(run_id,kind,message) VALUES($1,'queued','Investigasi masuk antrean persisten.')", []any{id}},
		{"INSERT INTO audit_events(case_id,actor,action,details) VALUES($1,$2,'investigation_queued',$3)", []any{caseID, actor, JSON(map[string]string{"run_id": id, "mode": mode})}},
	} {
		if _, e := tx.Exec(ctx, q.sql, q.args...); e != nil {
			return "", e
		}
	}
	return id, nil
}
func (s *Store) Enqueue(ctx context.Context, caseID, mode, instruction, actor string) (domain.Run, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return domain.Run{}, e
	}
	defer tx.Rollback(ctx)
	id, e := enqueue(ctx, tx, caseID, mode, instruction, actor)
	if e != nil {
		return domain.Run{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return domain.Run{}, e
	}
	return s.Run(ctx, id)
}
func (s *Store) Claim(ctx context.Context, conn *pgx.Conn) (domain.Run, error) {
	var id string
	e := conn.QueryRow(ctx, `WITH candidate AS(SELECT id FROM investigation_runs WHERE status='queued' OR(status='running' AND lease_until<now()) ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1)
 UPDATE investigation_runs r SET status='running',started_at=coalesce(started_at,now()),lease_until=now()+interval '60 seconds' FROM candidate c WHERE r.id=c.id RETURNING r.id`).Scan(&id)
	if e != nil {
		return domain.Run{}, normalize(e)
	}
	return s.Run(ctx, id)
}
func (s *Store) Heartbeat(ctx context.Context, id string) error {
	tag, e := s.Pool.Exec(ctx, "UPDATE investigation_runs SET lease_until=now()+interval '60 seconds' WHERE id=$1 AND status='running'", id)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrConflict
	}
	return e
}
func (s *Store) Checkpoint(ctx context.Context, id string, checkpoint any, rounds, tools int) error {
	tag, e := s.Pool.Exec(ctx, "UPDATE investigation_runs SET checkpoint=$2,model_rounds=$3,tool_calls=$4 WHERE id=$1 AND status='running'", id, JSON(checkpoint), rounds, tools)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrConflict
	}
	return e
}
func (s *Store) Finish(ctx context.Context, run domain.Run, status, message string, report *domain.Report) error {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var current string
	if e = tx.QueryRow(ctx, "SELECT status FROM investigation_runs WHERE id=$1 FOR UPDATE", run.ID).Scan(&current); e != nil {
		return e
	}
	if current == "cancelled" {
		return nil
	}
	if current != "running" {
		return ErrConflict
	}
	var raw any
	if report != nil {
		raw = JSON(report)
	}
	if _, e = tx.Exec(ctx, "UPDATE investigation_runs SET status=$2,error=$3,report=$4,finished_at=now(),lease_until=NULL WHERE id=$1", run.ID, status, message, raw); e != nil {
		return e
	}
	caseStatus := "open"
	if status == "succeeded" {
		caseStatus = "review_required"
		for _, a := range report.Recommendations {
			if _, e = tx.Exec(ctx, "INSERT INTO recommendations VALUES($1,$2,$3,$4,$5,$6,$7)", domain.ID(), run.ID, run.CaseID, a.Title, a.Rationale, JSON(a.EvidenceIDs), a.Department); e != nil {
				return e
			}
		}
	}
	if _, e = tx.Exec(ctx, "UPDATE anomaly_cases SET status=$2 WHERE id=$1", run.CaseID, caseStatus); e != nil {
		return e
	}
	kind := "completed"
	if status == "failed" {
		kind = "failed"
	}
	if message == "" {
		message = "Investigasi selesai; laporan tersedia untuk peninjauan."
	}
	if _, e = tx.Exec(ctx, "INSERT INTO investigation_events(run_id,kind,message) VALUES($1,$2,$3)", run.ID, kind, message); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO audit_events(case_id,actor,action,details) VALUES($1,'worker',$2,$3)", run.CaseID, "investigation_"+status, JSON(map[string]string{"run_id": run.ID})); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s *Store) Cancel(ctx context.Context, id, actor string) error {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var caseID string
	e = tx.QueryRow(ctx, "UPDATE investigation_runs SET status='cancelled',finished_at=now(),lease_until=NULL WHERE id=$1 AND status IN('queued','running') RETURNING case_id", id).Scan(&caseID)
	if errors.Is(e, pgx.ErrNoRows) {
		return ErrConflict
	}
	if e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "UPDATE anomaly_cases SET status='open' WHERE id=$1", caseID); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO investigation_events(run_id,kind,message) VALUES($1,'cancelled','Investigasi dibatalkan operator.')", id); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO audit_events(case_id,actor,action,details) VALUES($1,$2,'investigation_cancelled',$3)", caseID, actor, JSON(map[string]string{"run_id": id})); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s *Store) Review(ctx context.Context, id, decision, reason, actor string) error {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var caseID, runID string
	if e = tx.QueryRow(ctx, "SELECT case_id,run_id FROM recommendations WHERE id=$1", id).Scan(&caseID, &runID); e != nil {
		return normalize(e)
	}
	var locked string
	if e = tx.QueryRow(ctx, "SELECT id FROM anomaly_cases WHERE id=$1 FOR UPDATE", caseID).Scan(&locked); e != nil {
		return e
	}
	var latest, status string
	if e = tx.QueryRow(ctx, "SELECT id,status FROM investigation_runs WHERE case_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1", caseID).Scan(&latest, &status); e != nil {
		return e
	}
	if latest != runID || status != "succeeded" {
		return ErrConflict
	}
	tag, e := tx.Exec(ctx, "INSERT INTO review_actions VALUES($1,$2,$3,$4,$5,now()) ON CONFLICT(recommendation_id) DO NOTHING", domain.ID(), id, decision, reason, actor)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrConflict
	}
	if _, e = tx.Exec(ctx, "UPDATE anomaly_cases SET status=CASE WHEN EXISTS(SELECT 1 FROM recommendations r LEFT JOIN review_actions v ON v.recommendation_id=r.id WHERE r.run_id=$2 AND v.id IS NULL) THEN 'review_required' ELSE 'reviewed' END WHERE id=$1", caseID, runID); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO audit_events(case_id,actor,action,details) VALUES($1,$2,'recommendation_reviewed',$3)", caseID, actor, JSON(map[string]string{"recommendation_id": id, "decision": decision, "reason": reason})); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

type ReservationLimits struct {
	Total, Phase, Run int64
	PhaseName         string
}

func (s *Store) Reserve(ctx context.Context, runID string, maximum int64, l ReservationLimits) (string, error) {
	if maximum <= 0 {
		return "", fmt.Errorf("reservasi tidak valid")
	}
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return "", e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(2026102703)"); e != nil {
		return "", e
	}
	var total, phase, run int64
	if e = tx.QueryRow(ctx, "SELECT coalesce(sum(coalesce(actual_micros,reserved_micros)),0),coalesce(sum(coalesce(actual_micros,reserved_micros)) FILTER(WHERE phase=$1),0),coalesce(sum(coalesce(actual_micros,reserved_micros)) FILTER(WHERE run_id=$2),0) FROM budget_reservations", l.PhaseName, runID).Scan(&total, &phase, &run); e != nil {
		return "", e
	}
	if total+maximum > l.Total || phase+maximum > l.Phase || run+maximum > l.Run {
		return "", fmt.Errorf("budget_exhausted: reservasi melebihi batas kredit")
	}
	id := domain.ID()
	if _, e = tx.Exec(ctx, "INSERT INTO budget_reservations(id,run_id,phase,reserved_micros) VALUES($1,$2,$3,$4)", id, runID, l.PhaseName, maximum); e != nil {
		return "", e
	}
	if e = tx.Commit(ctx); e != nil {
		return "", e
	}
	return id, nil
}
func (s *Store) Settle(ctx context.Context, id string, input, output, actual int64) error {
	_, e := s.Pool.Exec(ctx, "UPDATE budget_reservations SET actual_micros=$2,input_tokens=$3,output_tokens=$4 WHERE id=$1 AND actual_micros IS NULL", id, actual, input, output)
	return e
}
