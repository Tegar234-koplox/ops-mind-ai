package store

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	migratedb "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"opsmind/internal/domain"
	"time"
)

//go:embed migrations/*.sql
var migrations embed.FS
var ErrConflict = errors.New("konflik state atau data usang")
var ErrNotFound = errors.New("record tidak ditemukan")

type Store struct{ Pool *pgxpool.Pool }

func Open(ctx context.Context, dsn string) (*Store, error) {
	c, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		return nil, e
	}
	c.MaxConns = 8
	c.MinConns = 0
	c.MaxConnIdleTime = time.Minute
	p, e := pgxpool.NewWithConfig(ctx, c)
	if e != nil {
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		p.Close()
		return nil, e
	}
	return &Store{p}, nil
}
func Migrate(dsn string) error {
	db, e := sql.Open("pgx", dsn)
	if e != nil {
		return e
	}
	defer db.Close()
	d, e := migratedb.WithInstance(db, &migratedb.Config{})
	if e != nil {
		return e
	}
	src, e := iofs.New(migrations, "migrations")
	if e != nil {
		return e
	}
	m, e := migrate.NewWithInstance("iofs", src, "pgx5", d)
	if e != nil {
		return e
	}
	defer m.Close()
	e = m.Up()
	if e == migrate.ErrNoChange {
		return nil
	}
	return e
}
func JSON(v any) []byte {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return b
}
func normalize(e error) error {
	if errors.Is(e, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return e
}

const caseCols = "id,rule_id,entity_id,shift_id,department,severity,status,data_quality,expected,observed,explanation,sources,detected_at"

func scanCase(row pgx.Row) (domain.Case, error) {
	var c domain.Case
	var a, b, s []byte
	e := row.Scan(&c.ID, &c.RuleID, &c.EntityID, &c.ShiftID, &c.Department, &c.Severity, &c.Status, &c.DataQuality, &a, &b, &c.Explanation, &s, &c.DetectedAt)
	if e != nil {
		return c, normalize(e)
	}
	for _, p := range []struct {
		b []byte
		v any
	}{{a, &c.Expected}, {b, &c.Observed}, {s, &c.Sources}} {
		if e = json.Unmarshal(p.b, p.v); e != nil {
			return c, e
		}
	}
	return c, nil
}
func (s *Store) Case(ctx context.Context, id string) (domain.Case, error) {
	return scanCase(s.Pool.QueryRow(ctx, "SELECT "+caseCols+" FROM anomaly_cases WHERE id=$1", id))
}
func (s *Store) Cases(ctx context.Context, limit, offset int, status, severity string) ([]domain.Case, int, error) {
	var total int
	e := s.Pool.QueryRow(ctx, "SELECT count(*) FROM anomaly_cases WHERE ($1='' OR status=$1) AND ($2='' OR severity=$2)", status, severity).Scan(&total)
	if e != nil {
		return nil, 0, e
	}
	rows, e := s.Pool.Query(ctx, "SELECT "+caseCols+" FROM anomaly_cases WHERE ($3='' OR status=$3) AND ($4='' OR severity=$4) ORDER BY CASE severity WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END,detected_at DESC,id LIMIT $1 OFFSET $2", limit, offset, status, severity)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	out := []domain.Case{}
	for rows.Next() {
		c, e := scanCase(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

const runCols = "id,case_id,status,mode,instruction,error,checkpoint,report,created_at,started_at,finished_at,model_rounds,tool_calls,(SELECT coalesce(sum(input_tokens),0) FROM budget_reservations b WHERE b.run_id=r.id),(SELECT coalesce(sum(output_tokens),0) FROM budget_reservations b WHERE b.run_id=r.id),(SELECT coalesce(sum(coalesce(actual_micros,reserved_micros)),0) FROM budget_reservations b WHERE b.run_id=r.id)"

func scanRun(row pgx.Row) (domain.Run, error) {
	var r domain.Run
	var raw []byte
	e := row.Scan(&r.ID, &r.CaseID, &r.Status, &r.Mode, &r.Instruction, &r.Error, &r.Checkpoint, &raw, &r.CreatedAt, &r.StartedAt, &r.FinishedAt, &r.ModelRounds, &r.ToolCalls, &r.InputTokens, &r.OutputTokens, &r.CostMicros)
	if e != nil {
		return r, normalize(e)
	}
	if raw != nil {
		e = json.Unmarshal(raw, &r.Report)
	}
	return r, e
}
func (s *Store) Run(ctx context.Context, id string) (domain.Run, error) {
	return scanRun(s.Pool.QueryRow(ctx, "SELECT "+runCols+" FROM investigation_runs r WHERE id=$1", id))
}
func (s *Store) LatestRun(ctx context.Context, id string) (domain.Run, error) {
	return scanRun(s.Pool.QueryRow(ctx, "SELECT "+runCols+" FROM investigation_runs r WHERE case_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1", id))
}
func (s *Store) Events(ctx context.Context, id string, after int64) ([]domain.Event, error) {
	rows, e := s.Pool.Query(ctx, "SELECT id,run_id,kind,message,created_at FROM investigation_events WHERE run_id=$1 AND id>$2 ORDER BY id LIMIT 500", id, after)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	a := []domain.Event{}
	for rows.Next() {
		var v domain.Event
		if e = rows.Scan(&v.ID, &v.RunID, &v.Kind, &v.Message, &v.CreatedAt); e != nil {
			return nil, e
		}
		a = append(a, v)
	}
	return a, rows.Err()
}
func (s *Store) Event(ctx context.Context, id, kind, message string) error {
	_, e := s.Pool.Exec(ctx, "INSERT INTO investigation_events(run_id,kind,message) VALUES($1,$2,$3)", id, kind, message)
	return e
}
func (s *Store) Evidence(ctx context.Context, caseID, id string) (domain.Evidence, error) {
	var v domain.Evidence
	var f, src, n []byte
	e := s.Pool.QueryRow(ctx, "SELECT id,case_id,run_id,shift_id,tool,domain,facts,sources,completeness,notes,hash,retrieved_at FROM case_evidence WHERE case_id=$1 AND id=$2", caseID, id).Scan(&v.ID, &v.CaseID, &v.RunID, &v.ShiftID, &v.Tool, &v.Domain, &f, &src, &v.Completeness, &n, &v.Hash, &v.RetrievedAt)
	if e != nil {
		return v, normalize(e)
	}
	if e = json.Unmarshal(f, &v.Facts); e != nil {
		return v, e
	}
	if e = json.Unmarshal(src, &v.Sources); e != nil {
		return v, e
	}
	e = json.Unmarshal(n, &v.Notes)
	return v, e
}
func (s *Store) EvidenceList(ctx context.Context, runID string) ([]domain.Evidence, error) {
	rows, e := s.Pool.Query(ctx, "SELECT case_id,id FROM case_evidence WHERE run_id=$1 ORDER BY retrieved_at,id", runID)
	if e != nil {
		return nil, e
	}
	type pair struct{ c, id string }
	p := []pair{}
	for rows.Next() {
		var v pair
		if e = rows.Scan(&v.c, &v.id); e != nil {
			rows.Close()
			return nil, e
		}
		p = append(p, v)
	}
	err := rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	out := []domain.Evidence{}
	for _, v := range p {
		a, e := s.Evidence(ctx, v.c, v.id)
		if e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, nil
}
func (s *Store) Recommendations(ctx context.Context, caseID string) ([]domain.Recommendation, error) {
	rows, e := s.Pool.Query(ctx, "SELECT x.id,x.run_id,x.case_id,x.title,x.rationale,x.evidence_ids,x.department,coalesce(v.decision,''),coalesce(v.reason,''),coalesce(v.reviewer,'') FROM recommendations x LEFT JOIN review_actions v ON v.recommendation_id=x.id WHERE x.case_id=$1 ORDER BY x.id", caseID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Recommendation{}
	for rows.Next() {
		var r domain.Recommendation
		var raw []byte
		if e = rows.Scan(&r.ID, &r.RunID, &r.CaseID, &r.Title, &r.Rationale, &raw, &r.Department, &r.Decision, &r.Reason, &r.Reviewer); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(raw, &r.EvidenceIDs); e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (s *Store) Audit(ctx context.Context, caseID string) ([]map[string]any, error) {
	rows, e := s.Pool.Query(ctx, "SELECT id,actor,action,details,created_at FROM audit_events WHERE case_id=$1 ORDER BY id", caseID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	a := []map[string]any{}
	for rows.Next() {
		var id int64
		var actor, action string
		var raw []byte
		var at time.Time
		if e = rows.Scan(&id, &actor, &action, &raw, &at); e != nil {
			return nil, e
		}
		var details any
		if e = json.Unmarshal(raw, &details); e != nil {
			return nil, e
		}
		a = append(a, map[string]any{"id": id, "actor": actor, "action": action, "details": details, "created_at": at})
	}
	return a, rows.Err()
}
func (s *Store) Budget(ctx context.Context) (map[string]int64, error) {
	v := map[string]int64{"total": 0, "development": 0, "demo": 0}
	rows, e := s.Pool.Query(ctx, "SELECT phase,coalesce(sum(coalesce(actual_micros,reserved_micros)),0) FROM budget_reservations GROUP BY phase")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var p string
		var n int64
		if e = rows.Scan(&p, &n); e != nil {
			return nil, e
		}
		v[p] = n
		v["total"] += n
	}
	return v, rows.Err()
}
func safeErr(e error) error { return fmt.Errorf("database: %w", e) }
