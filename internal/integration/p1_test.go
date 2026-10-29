package integration

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"opsmind/internal/agent"
	"opsmind/internal/config"
	"opsmind/internal/detection"
	"opsmind/internal/domain"
	"opsmind/internal/seed"
	"opsmind/internal/store"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestP1RulesAndEvidence(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL belum diisi")
	}
	u, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("gunakan database berakhiran _test")
	}
	ctx := context.Background()
	s, e := store.Open(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Pool.Close()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, e := s.Pool.Exec(ctx, q, args...); e != nil {
			t.Fatal(e)
		}
	}
	exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public")
	if e = store.Migrate(dsn); e != nil {
		t.Fatal(e)
	}
	// Upgrade a populated v3 dataset without inventing confirmations or closing invoices.
	for _, file := range []string{"000005_shift_handovers.down.sql", "000004_p1_sources.down.sql"} {
		down, err := os.ReadFile(filepath.Join("..", "store", "migrations", file))
		if err != nil {
			t.Fatal(err)
		}
		exec(string(down))
	}
	exec("UPDATE schema_migrations SET version=3")
	s.Pool.Reset()
	if _, e = seed.Run(ctx, s); e != nil {
		t.Fatal(e)
	}
	if e = store.Migrate(dsn); e != nil {
		t.Fatal(e)
	}
	s.Pool.Reset()
	sh, e := s.Shift(ctx, seed.GoldenShift)
	if e != nil {
		t.Fatal(e)
	}
	v, e := s.ExpenseVariance(ctx, sh, domain.AsOf)
	if e != nil || v.Complete {
		t.Fatal("upgrade fabricated posting closure", v, e)
	}
	if created, e := seed.RunP1(ctx, s); e != nil || !created {
		t.Fatal("seed P1", created, e)
	}
	if created, e := seed.RunP1(ctx, s); e != nil || created {
		t.Fatal("P1 seed must be idempotent", created, e)
	}
	// Leave a clean deterministic dataset for the following browser step in CI.
	t.Cleanup(func() {
		s.Pool.Close()
		a, err := store.Open(ctx, dsn)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = a.Pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public")
		a.Pool.Close()
		if err != nil {
			t.Error(err)
			return
		}
		if err = store.Migrate(dsn); err != nil {
			t.Error(err)
			return
		}
		fresh, err := store.Open(ctx, dsn)
		if err != nil {
			t.Error(err)
			return
		}
		defer fresh.Pool.Close()
		if _, err = seed.Run(ctx, fresh); err != nil {
			t.Error(err)
			return
		}
		if _, err = seed.RunP1(ctx, fresh); err != nil {
			t.Error(err)
		}
	})
	cfg := config.Config{Mode: "mock", EnableP1: true, CostThreshold: 2000, OTThreshold: 2000, ProdThreshold: 1000, DowntimeThreshold: 60, AttendanceGrace: 15, MaxRounds: 8, MaxTools: 16, MaxRetries: 2, RunTimeout: 180 * time.Second, ModelTimeout: 30 * time.Second, CreditLimit: 20000000, DevelopmentLimit: 15000000, DemoLimit: 5000000, RunLimit: 250000, Phase: "development"}
	engine := detection.Engine{Store: s, Config: cfg}
	found := map[string]domain.Case{}
	for _, id := range []string{seed.P1AttendanceShift, seed.P1OvertimeShift, seed.P1CostHighShift, seed.P1CostLowShift} {
		cases, e := engine.Execute(ctx, id, domain.AsOf)
		if e != nil {
			t.Fatal(id, e)
		}
		for _, c := range cases {
			found[c.RuleID] = c
		}
	}
	for _, rule := range []string{"ATT-001", "OT-002", "COST-001", "COST-002"} {
		c, ok := found[rule]
		if !ok || c.DataQuality != "confirmed" {
			t.Fatal("missing P1 rule", rule, c)
		}
		if e = s.ValidateCaseSources(ctx, c, c.Sources); e != nil {
			t.Fatal(rule, e)
		}
	}
	golden, e := engine.Execute(ctx, seed.GoldenShift, domain.AsOf)
	if e != nil || len(golden) != 4 {
		t.Fatal("golden changed", len(golden), e)
	}
	t.Run("confirmed_streak_and_exceptions", func(t *testing.T) {
		sh, _ := s.Shift(ctx, seed.P1AttendanceShift)
		v, e := s.AbsenceStreaks(ctx, sh, domain.AsOf)
		if e != nil || len(v.Employees) != 1 {
			t.Fatal(v, e)
		}
		staff, e := s.Staffing(ctx, sh.ID)
		if e != nil || staff.ConfirmedAbsent != 1 || staff.Unverified != 0 {
			t.Fatal(staff, e)
		}
		var aid, employee string
		var begin, end, confirmed time.Time
		if e = s.Pool.QueryRow(ctx, "SELECT a.id,a.employee_id,sh.starts_at,sh.ends_at,t.absence_confirmed_at FROM shift_assignments a JOIN shift_instances sh ON sh.id=a.shift_id JOIN attendance_records t ON t.assignment_id=a.id WHERE a.shift_id=$1 AND t.absence_confirmed_at IS NOT NULL", sh.ID).Scan(&aid, &employee, &begin, &end, &confirmed); e != nil {
			t.Fatal(e)
		}
		if _, err := s.Pool.Exec(ctx, "UPDATE attendance_records SET absence_confirmed_by=NULL WHERE assignment_id=$1", aid); err == nil {
			t.Fatal("confirmed absence accepted without actor")
		}
		expectNone := func() {
			t.Helper()
			v, e := s.AbsenceStreaks(ctx, sh, domain.AsOf)
			if e != nil || len(v.Employees) != 0 {
				t.Fatal(v, e)
			}
			ev, err := agent.ExecuteTool(ctx, s, found["ATT-001"], domain.Run{ID: "p1-stale-run"}, "get_attendance_summary", agent.ToolArgs{ShiftID: sh.ID})
			if err != nil || ev.Facts["consecutive_shifts"] != "0" || ev.Completeness != "needs_verification" {
				t.Fatal("stale attendance case retained confirmed streak", ev, err)
			}
		}
		exec("INSERT INTO leave_records VALUES('P1-TEST-LEAVE',$1,$2,$3,true)", employee, begin, end)
		expectNone()
		exec("DELETE FROM leave_records WHERE id='P1-TEST-LEAVE'")
		exec("UPDATE attendance_records SET absence_confirmed_at=$2 WHERE assignment_id=$1", aid, domain.AsOf.Add(time.Minute))
		expectNone()
		exec("UPDATE attendance_records SET absence_confirmed_at=NULL,absence_confirmed_by=NULL,source_complete=false WHERE assignment_id=$1", aid)
		expectNone()
		exec("UPDATE attendance_records SET source_complete=true,absence_confirmed_at=$2,absence_confirmed_by='SUP-SYNTHETIC' WHERE assignment_id=$1", aid, confirmed)
		ev, e := agent.ExecuteTool(ctx, s, found["ATT-001"], domain.Run{ID: "p1-streak-run"}, "get_attendance_summary", agent.ToolArgs{ShiftID: sh.ID})
		if e != nil || ev.Facts["confirmed_absence_workers"] != "1" {
			t.Fatal(ev, e)
		}
		if e = s.ValidateCaseSources(ctx, found["ATT-001"], []domain.SourceRef{{Type: "attendance_records", ID: "AT-AS-" + seed.GoldenShift + "-00"}}); e == nil {
			t.Fatal("unrelated attendance accepted")
		}
		disabled := engine
		disabled.Config.EnableP1 = false
		cases, e := disabled.Execute(ctx, sh.ID, domain.AsOf)
		if e != nil {
			t.Fatal(e)
		}
		for _, c := range cases {
			if c.RuleID == "ATT-001" {
				t.Fatal("P1 flag ignored")
			}
		}
	})
	t.Run("approval_revocation_and_future_events", func(t *testing.T) {
		sh, _ := s.Shift(ctx, seed.P1OvertimeShift)
		a, e := s.OvertimeApprovals(ctx, sh, domain.AsOf)
		if e != nil || a.Unapproved != 1 || a.Minutes <= 0 {
			t.Fatal(a, e)
		}
		if _, e = s.Pool.Exec(ctx, "UPDATE overtime_approval_events SET status='approved'"); e == nil {
			t.Fatal("approval history mutable")
		}
		var oid string
		if e = s.Pool.QueryRow(ctx, "SELECT overtime_id FROM overtime_approval_events WHERE id='P1-OT-REVOKED'").Scan(&oid); e != nil {
			t.Fatal(e)
		}
		exec("INSERT INTO overtime_approval_events VALUES('P1-TEST-REAPPROVED',$1,'approved',$2,'SUP-TEST')", oid, domain.AsOf.Add(-time.Minute))
		exec("INSERT INTO overtime_approval_events VALUES('P1-TEST-FUTURE',$1,'revoked',$2,'SUP-TEST')", oid, domain.AsOf.Add(time.Minute))
		a, e = s.OvertimeApprovals(ctx, sh, domain.AsOf)
		if e != nil || a.Unapproved != 0 {
			t.Fatal("future revocation applied", a, e)
		}
		exec("INSERT INTO overtime_approval_events VALUES('P1-TEST-REVOKED',$1,'revoked',$2,'SUP-TEST')", oid, domain.AsOf.Add(-30*time.Second))
		exec("UPDATE overtime_records SET complete=false WHERE id=$1", oid)
		a, e = s.OvertimeApprovals(ctx, sh, domain.AsOf)
		if e != nil || a.Unapproved != 0 || a.Unverified != 1 || a.PendingApproval != 1 {
			t.Fatal("incomplete overtime confirmed", a, e)
		}
		exec("UPDATE overtime_records SET complete=true WHERE id=$1", oid)
		// An incomplete unapproved source must create a verification case, not a confirmed one.
		var pendingID string
		if e = s.Pool.QueryRow(ctx, "SELECT o.id FROM overtime_records o JOIN shift_assignments a ON a.id=o.assignment_id WHERE a.shift_id=$1 ORDER BY o.id LIMIT 1", seed.P1CostLowShift).Scan(&pendingID); e != nil {
			t.Fatal(e)
		}
		exec("UPDATE overtime_records SET complete=false,approved=false WHERE id=$1", pendingID)
		pendingCases, err := engine.Execute(ctx, seed.P1CostLowShift, domain.AsOf)
		if err != nil {
			t.Fatal(err)
		}
		needsVerification := false
		for _, c := range pendingCases {
			if c.RuleID == "OT-002" {
				needsVerification = c.DataQuality == "needs_verification" && c.Observed["unverified_unapproved_records"] == "1"
			}
		}
		if !needsVerification {
			t.Fatal("incomplete approval must stay needs_verification")
		}
		exec("UPDATE overtime_records SET complete=true,approved=true WHERE id=$1", pendingID)
	})
	t.Run("cost_thresholds_closure_baseline_and_scope", func(t *testing.T) {
		sh, _ := s.Shift(ctx, seed.P1CostHighShift)
		// Normalize all four comparison ledgers to an exact 500000 to exercise strict +/-20% boundaries.
		exec("UPDATE expenses SET amount_rupiah=200000 WHERE category='maintenance' AND shift_id IN(SELECT shift_id FROM expense_closures WHERE shift_id<>$1 AND shift_id LIKE 'B-%')", sh.ID)
		contains := func(rule string) bool {
			t.Helper()
			cases, e := engine.Execute(ctx, sh.ID, domain.AsOf)
			if e != nil {
				t.Fatal(e)
			}
			for _, c := range cases {
				if c.RuleID == rule {
					return true
				}
			}
			return false
		}
		for _, tc := range []struct {
			total int64
			rule  string
			want  bool
		}{{600000, "COST-001", false}, {600100, "COST-001", true}, {400000, "COST-002", false}, {399900, "COST-002", true}} {
			exec("UPDATE expenses SET amount_rupiah=$2 WHERE shift_id=$1 AND category='maintenance'", sh.ID, tc.total-300000)
			if got := contains(tc.rule); got != tc.want {
				t.Fatalf("total %d %s got %v", tc.total, tc.rule, got)
			}
		}
		v, e := s.ExpenseVariance(ctx, sh, domain.AsOf)
		if e != nil || v.Baseline != 500000 || !v.Complete {
			t.Fatal(v, e)
		}
		exec("UPDATE expense_closures SET closed_at=$2 WHERE shift_id=$1", sh.ID, sh.Start)
		if contains("COST-002") {
			t.Fatal("closure before shift end accepted")
		}
		exec("UPDATE expense_closures SET closed_at=$2 WHERE shift_id=$1", sh.ID, domain.AsOf.Add(time.Minute))
		if contains("COST-002") {
			t.Fatal("future closure accepted")
		}
		exec("UPDATE expense_closures SET closed_at=$2 WHERE shift_id=$1", sh.ID, sh.End.Add(5*time.Minute))
		exec("UPDATE expenses SET complete=false WHERE shift_id=$1 AND category='maintenance'", sh.ID)
		if contains("COST-002") {
			t.Fatal("missing invoice treated as underspend")
		}
		exec("UPDATE expenses SET complete=true WHERE shift_id=$1", sh.ID)
		var baselineID string
		if e = s.Pool.QueryRow(ctx, "SELECT shift_id FROM expense_closures WHERE shift_id LIKE 'B-%' AND shift_id<>$1 ORDER BY shift_id LIMIT 1", sh.ID).Scan(&baselineID); e != nil {
			t.Fatal(e)
		}
		exec("UPDATE expense_closures SET closed_at=$2 WHERE shift_id=$1", baselineID, domain.AsOf.Add(time.Minute))
		v, e = s.ExpenseVariance(ctx, sh, domain.AsOf)
		if e != nil || v.Complete || v.Comparable != 3 {
			t.Fatal("insufficient baseline", v, e)
		}
		exec("UPDATE expense_closures SET closed_at=(SELECT ends_at+interval '5 minutes' FROM shift_instances WHERE id=$1) WHERE shift_id=$1", baselineID)
		ev, e := agent.ExecuteTool(ctx, s, found["COST-001"], domain.Run{ID: "p1-cost-run"}, "get_cost_breakdown", agent.ToolArgs{ShiftID: sh.ID})
		if e != nil || ev.Facts["expense_baseline_rupiah"] != "500000" {
			t.Fatal(ev, e)
		}
		if e = s.ValidateCaseSources(ctx, found["COST-001"], []domain.SourceRef{{Type: "expenses", ID: "EX-MT-" + seed.P1CostLowShift}}); e == nil {
			t.Fatal("foreign expense accepted")
		}
		exec("UPDATE expenses SET amount_rupiah=0 WHERE shift_id IN(SELECT shift_id FROM expense_closures WHERE shift_id LIKE 'B-%' AND shift_id<>$1)", sh.ID)
		v, e = s.ExpenseVariance(ctx, sh, domain.AsOf)
		if e != nil || v.Complete {
			t.Fatal("zero baseline accepted", v, e)
		}
	})
	// Execute a full P1 investigation through the unchanged eight-tool contract.
	run, e := s.Enqueue(ctx, found["ATT-001"].ID, "mock", "Uji P1 sintetis", "operator")
	if e != nil {
		t.Fatal(e)
	}
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		(&agent.Runner{Store: s, Config: cfg, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Work(workerCtx)
	}()
	run = awaitRun(t, s, run.ID)
	cancel()
	<-done
	if run.Status != "succeeded" {
		t.Fatal("P1 agent", run.Status, run.Error)
	}
}
