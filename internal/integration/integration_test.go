package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"opsmind/internal/agent"
	"opsmind/internal/config"
	"opsmind/internal/detection"
	"opsmind/internal/domain"
	"opsmind/internal/httpapi"
	"opsmind/internal/seed"
	"opsmind/internal/store"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoldenWorkflow(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL belum diisi; pengujian PostgreSQL tidak dijalankan")
	}
	u, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("gunakan database khusus berakhiran _test")
	}
	ctx := context.Background()
	s, e := store.Open(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Pool.Close()
	if _, e = s.Pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); e != nil {
		t.Fatal(e)
	}
	if e = store.Migrate(dsn); e != nil {
		t.Fatal(e)
	}
	if e = store.Migrate(dsn); e != nil {
		t.Fatal("repeat migration", e)
	}
	if _, e = seed.Run(ctx, s); e != nil {
		t.Fatal(e)
	}
	if created, e := seed.Run(ctx, s); e != nil || created {
		t.Fatal("seed idempotency", created, e)
	}
	var employees, overlap, days, machines int
	if e = s.Pool.QueryRow(ctx, "SELECT count(*) FROM employees").Scan(&employees); e != nil || employees != 150 {
		t.Fatal(employees, e)
	}
	if e = s.Pool.QueryRow(ctx, "SELECT count(*) FROM calendar_days").Scan(&days); e != nil || days != 60 {
		t.Fatal(days, e)
	}
	if e = s.Pool.QueryRow(ctx, "SELECT count(*) FROM machines").Scan(&machines); e != nil || machines != 8 {
		t.Fatal(machines, e)
	}
	if e = s.Pool.QueryRow(ctx, "SELECT count(*) FROM shift_assignments a JOIN shift_assignments b ON b.employee_id=a.employee_id AND a.id<b.id JOIN shift_instances x ON x.id=a.shift_id JOIN shift_instances y ON y.id=b.shift_id WHERE x.starts_at<y.ends_at AND y.starts_at<x.ends_at").Scan(&overlap); e != nil || overlap != 0 {
		t.Fatal("overlapping employee assignments", overlap, e)
	}
	cfg := config.Config{DSN: dsn, Origin: "http://opsmind.test", Username: "operator", Password: "integration-test-password", SessionSecret: strings.Repeat("s", 40), Mode: "mock", MaxRounds: 8, MaxTools: 16, MaxRetries: 2, RunTimeout: 180 * time.Second, ModelTimeout: 30 * time.Second, OTThreshold: 2000, ProdThreshold: 1000, DowntimeThreshold: 60, AttendanceGrace: 15, CreditLimit: 20000000, DevelopmentLimit: 15000000, DemoLimit: 5000000, RunLimit: 250000, Phase: "development", WebDir: "nonexistent"}
	engine := detection.Engine{Store: s, Config: cfg}
	cases, e := engine.Execute(ctx, seed.GoldenShift, domain.AsOf)
	if e != nil {
		t.Fatal(e)
	}
	if len(cases) != 4 {
		t.Fatalf("expected four rules got %d", len(cases))
	}
	var primary domain.Case
	for _, c := range cases {
		switch c.RuleID {
		case "OT-001":
			primary = c
			if c.Observed["variance_bps"] != "3800" || c.Observed["overtime_rupiah"] != "1380000" {
				t.Fatal(c)
			}
		case "PROD-001":
			if c.Observed["shortfall_bps"] != "1700" {
				t.Fatal(c)
			}
		case "ATT-002":
			if c.Observed["unverified"] != "4" || c.Observed["approved_leave"] != "3" || c.Observed["verified_present"] != "31" {
				t.Fatal(c)
			}
		case "MAINT-001":
			if c.Observed["downtime_minutes"] != "110" {
				t.Fatal(c)
			}
		}
	}
	if _, e = engine.Execute(ctx, seed.GoldenShift, domain.AsOf); e != nil {
		t.Fatal(e)
	}
	var caseCount int
	if e = s.Pool.QueryRow(ctx, "SELECT count(*) FROM anomaly_cases").Scan(&caseCount); e != nil || caseCount != 4 {
		t.Fatal("duplicate detection", caseCount, e)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	workerCtx, stopWorker := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); (&agent.Runner{Store: s, Config: cfg, Logger: log}).Work(workerCtx) }()
	run, e := s.LatestRun(ctx, primary.ID)
	if e != nil {
		t.Fatal(e)
	}
	run = awaitRun(t, s, run.ID)
	stopWorker()
	<-done
	if run.Status != "succeeded" || run.ToolCalls != 8 {
		t.Fatalf("run %+v", run)
	}
	evidence, e := s.EvidenceList(ctx, run.ID)
	if e != nil || len(evidence) != 7 {
		t.Fatalf("evidence %d %v", len(evidence), e)
	}
	if !agent.HasDomains(evidence) {
		t.Fatal("domain coverage")
	}
	for _, ev := range evidence {
		if e = s.ValidateSources(ctx, seed.GoldenShift, ev.Sources); e != nil {
			t.Fatal(e)
		}
		if ev.Hash != agent.EvidenceHash(ev) {
			t.Fatal("snapshot hash")
		}
	}
	if _, e = agent.ValidateReport(store.JSON(run.Report), evidence); e != nil {
		t.Fatal(e)
	}
	recs, e := s.Recommendations(ctx, primary.ID)
	if e != nil || len(recs) != 2 {
		t.Fatal(len(recs), e)
	}
	// Auth, origin, idempotency, review, and data-scope checks through real HTTP handlers.
	api, e := httpapi.New(s, cfg, log)
	if e != nil {
		t.Fatal(e)
	}
	ts := httptest.NewServer(api)
	defer ts.Close()
	cookie := ""
	request := func(method, p, body, key, origin string) (int, []byte) {
		t.Helper()
		req, e := http.NewRequest(method, ts.URL+"/api/v1"+p, bytes.NewBufferString(body))
		if e != nil {
			t.Fatal(e)
		}
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if p == "/auth/login" && resp.StatusCode == 200 {
			cookie = resp.Header.Get("Set-Cookie")
		}
		return resp.StatusCode, raw
	}
	if status, _ := request("GET", "/cases", "", "", ""); status != 401 {
		t.Fatal(status)
	}
	if status, _ := request("POST", "/auth/login", `{"username":"operator","password":"integration-test-password"}`, "", cfg.Origin); status != 200 {
		t.Fatal("login", status)
	}
	reviewPath := "/recommendations/" + recs[0].ID + "/reviews"
	body := `{"decision":"approved","reason":"Ditinjau supervisor"}`
	if status, _ := request("POST", reviewPath, body, "origin-test", "http://evil.test"); status != 403 {
		t.Fatal("origin", status)
	}
	status, first := request("POST", reviewPath, body, "review-once", cfg.Origin)
	if status != 200 {
		t.Fatal("review", status, string(first))
	}
	status, replayed := request("POST", reviewPath, body, "review-once", cfg.Origin)
	if status != 200 || !bytes.Equal(first, replayed) {
		t.Fatal("idempotent replay", status)
	}
	if status, _ = request("POST", reviewPath, `{"decision":"rejected","reason":"x"}`, "review-once", cfg.Origin); status != 409 {
		t.Fatal("idem body conflict", status)
	}
	if status, _ = request("POST", reviewPath, body, "review-again", cfg.Origin); status != 409 {
		t.Fatal("double review", status)
	}
	if status, _ = request("POST", "/recommendations/"+recs[1].ID+"/reviews", `{"decision":"rejected","reason":""}`, "reject-empty", cfg.Origin); status != 400 {
		t.Fatal("reject without reason", status)
	}
	if status, _ = request("POST", "/recommendations/"+recs[1].ID+"/reviews", `{"decision":"rejected","reason":"Perlu log tambahan"}`, "reject-valid", cfg.Origin); status != 200 {
		t.Fatal("reject", status)
	}
	if status, _ = request("GET", "/cases/foreign/evidence/"+evidence[0].ID, "", "", ""); status != 404 {
		t.Fatal("foreign evidence", status)
	}
	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/investigations/"+run.ID+"/events", nil)
	req.Header.Set("Cookie", cookie)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	stream, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Contains(stream, []byte("event: progress")) {
		t.Fatal("SSE missing")
	}
	events, e := s.Events(ctx, run.ID, 0)
	if e != nil || len(events) < 2 {
		t.Fatal(e)
	}
	cursor := events[len(events)-2].ID
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/investigations/%s/events?after=%d", ts.URL, run.ID, cursor), nil)
	req.Header.Set("Cookie", cookie)
	resp, e = http.DefaultClient.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	resumed, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if bytes.Count(resumed, []byte("event: progress")) != 1 {
		t.Fatal("SSE replay duplicated")
	}
	// Atomic budget limits across concurrent callers; unresolved requests retain reservations.
	var ok atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Reserve(ctx, run.ID, 25000, store.ReservationLimits{Total: 50000, Phase: 50000, Run: 50000, PhaseName: "development"})
			if err == nil {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 2 {
		t.Fatalf("budget race: %d accepted", ok.Load())
	}
	if _, e = s.Pool.Exec(ctx, "UPDATE audit_events SET actor='tamper'"); e == nil {
		t.Fatal("audit mutable")
	}
	// Interrupted worker resumes pending transcript without duplicating successful tool execution.
	next, e := s.Enqueue(ctx, primary.ID, "mock", "Periksa kembali bukti", "operator")
	if e != nil {
		t.Fatal(e)
	}
	cancelled, e := s.Enqueue(ctx, cases[1].ID, "mock", "", "operator")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Cancel(ctx, cancelled.ID, "operator"); e != nil {
		t.Fatal(e)
	}
	interruptCtx, interrupt := context.WithCancel(ctx)
	interrupted := make(chan struct{})
	go func() {
		defer close(interrupted)
		(&agent.Runner{Store: s, Config: cfg, Logger: log, Model: &interruptModel{cancel: interrupt}}).Work(interruptCtx)
	}()
	select {
	case <-interrupted:
	case <-time.After(15 * time.Second):
		t.Fatal("interruption failed")
	}
	var saved int
	if e = s.Pool.QueryRow(ctx, "SELECT count(*) FROM tool_executions WHERE run_id=$1", next.ID).Scan(&saved); e != nil || saved != 2 {
		t.Fatal("checkpoint lost", saved, e)
	}
	if _, e = s.Pool.Exec(ctx, "UPDATE investigation_runs SET lease_until=now()-interval '1 second' WHERE id=$1", next.ID); e != nil {
		t.Fatal(e)
	}
	resumeCtx, resumeStop := context.WithCancel(ctx)
	resumeDone := make(chan struct{})
	go func() { defer close(resumeDone); (&agent.Runner{Store: s, Config: cfg, Logger: log}).Work(resumeCtx) }()
	next = awaitRun(t, s, next.ID)
	resumeStop()
	<-resumeDone
	if next.Status != "succeeded" || next.ToolCalls != 8 {
		t.Fatal("resume", next)
	}
	var calls int
	if e = s.Pool.QueryRow(ctx, "SELECT count(*) FROM tool_executions WHERE run_id=$1", next.ID).Scan(&calls); e != nil || calls != 8 {
		t.Fatal("duplicate tool calls", calls, e)
	}
	if e = s.Review(ctx, recs[0].ID, "approved", "", "operator"); e != store.ErrConflict {
		t.Fatal("old recommendation accepted", e)
	}
}

type interruptModel struct {
	calls  int
	cancel context.CancelFunc
}

func (m *interruptModel) Complete(ctx context.Context, r domain.Run, messages []agent.Message) (agent.Message, error) {
	m.calls++
	if m.calls == 2 {
		m.cancel()
		return agent.Message{}, context.Canceled
	}
	return (agent.MockModel{}).Complete(ctx, r, messages)
}
func awaitRun(t *testing.T, s *store.Store, id string) domain.Run {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		run, e := s.Run(context.Background(), id)
		if e != nil {
			t.Fatal(e)
		}
		if domain.Terminal(run.Status) {
			return run
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("run timeout")
	return domain.Run{}
}
