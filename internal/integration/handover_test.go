package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"opsmind/internal/agent"
	"opsmind/internal/config"
	"opsmind/internal/detection"
	"opsmind/internal/domain"
	"opsmind/internal/handover"
	"opsmind/internal/httpapi"
	"opsmind/internal/seed"
	"opsmind/internal/store"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHandoverSnapshotsAndAPI(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL belum diisi")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("gunakan database khusus berakhiran _test")
	}
	ctx := context.Background()
	s, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Pool.Close()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := s.Pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public")
	if err = store.Migrate(dsn); err != nil {
		t.Fatal(err)
	}
	if _, err = seed.Run(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err = seed.RunP1(ctx, s); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Origin: "http://opsmind.test", Username: "operator", Password: "handover-test-password", SessionSecret: strings.Repeat("s", 40), Mode: "mock", MaxRounds: 8, MaxTools: 16, MaxRetries: 2, RunTimeout: 180 * time.Second, ModelTimeout: 30 * time.Second, OTThreshold: 2000, ProdThreshold: 1000, DowntimeThreshold: 60, AttendanceGrace: 15, CreditLimit: 20000000, DevelopmentLimit: 15000000, DemoLimit: 5000000, RunLimit: 250000, Phase: "development", EnableP1: true, CostThreshold: 2000}
	cases, err := (detection.Engine{Store: s, Config: cfg}).Execute(ctx, seed.GoldenShift, domain.AsOf)
	if err != nil || len(cases) != 4 {
		t.Fatal(cases, err)
	}
	var primary domain.Case
	for _, c := range cases {
		if c.RuleID == "OT-001" {
			primary = c
		}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	workerCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); (&agent.Runner{Store: s, Config: cfg, Logger: log}).Work(workerCtx) }()
	run, err := s.LatestRun(ctx, primary.ID)
	if err != nil {
		stop()
		<-done
		t.Fatal(err)
	}
	run = awaitRun(t, s, run.ID)
	stop()
	<-done
	if run.Status != "succeeded" {
		t.Fatal(run.Status)
	}
	recs, err := s.Recommendations(ctx, primary.ID)
	if err != nil || len(recs) < 2 {
		t.Fatal(recs, err)
	}
	if err = s.Review(ctx, recs[0].ID, "approved", "Uji serah terima sintetis", "operator"); err != nil {
		t.Fatal(err)
	}
	var runsBefore, budgetBefore int
	if err = s.Pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM investigation_runs),(SELECT count(*) FROM budget_reservations)").Scan(&runsBefore, &budgetBefore); err != nil {
		t.Fatal(err)
	}
	service := handover.Service{Store: s}
	first, err := service.Create(ctx, seed.GoldenShift, "operator")
	if err != nil {
		t.Fatal(err)
	}
	facts := map[string]map[string]string{}
	for _, section := range first.Snapshot.Sections {
		facts[section.Key] = section.Facts
	}
	if facts["workforce"]["verified_present"] != "31" || facts["workforce"]["approved_leave"] != "3" || facts["workforce"]["unverified"] != "4" || facts["finance"]["variance_bps"] != "3800" || facts["production"]["shortfall_bps"] != "1700" || facts["maintenance"]["downtime_minutes"] != "110" {
		t.Fatal(facts)
	}
	if first.Snapshot.NextShift == nil || first.Snapshot.NextShift.ID != "A-20261026-S2" {
		t.Fatal(first.Snapshot.NextShift)
	}
	if len(first.Snapshot.Cases) != 4 || first.Snapshot.Cases[0].ModelMode != "mock" {
		t.Fatal(first.Snapshot.Cases)
	}
	if first.Snapshot.Cases[0].Recommendations[0].Decision != "approved" {
		t.Fatal("review missing")
	}
	if err = s.Review(ctx, recs[1].ID, "rejected", "Perlu verifikasi tambahan", "operator"); err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(ctx, seed.GoldenShift, "operator")
	if err != nil {
		t.Fatal(err)
	}
	if second.Hash == first.Hash {
		t.Fatal("new review was not captured")
	}
	saved, err := s.Handover(ctx, first.ID)
	if err != nil || saved.Hash != first.Hash || saved.Snapshot.Cases[0].Recommendations[1].Decision != "" {
		t.Fatal("old snapshot changed", err)
	}
	if _, err = s.Pool.Exec(ctx, "UPDATE shift_handovers SET created_by='tamper' WHERE id=$1", first.ID); err == nil {
		t.Fatal("handover mutable")
	}
	if _, err = s.Pool.Exec(ctx, "DELETE FROM shift_handovers WHERE id=$1", first.ID); err == nil {
		t.Fatal("handover deletable")
	}
	tampered := first
	tampered.ID = domain.ID()
	tampered.Hash = strings.Repeat("0", 64)
	if err = s.SaveHandover(ctx, tampered); err == nil {
		t.Fatal("tampered snapshot saved")
	}
	if _, err = service.Create(ctx, "missing", "operator"); !errors.Is(err, store.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err = s.HandoverData(ctx, seed.GoldenShift, domain.AsOf.Add(-24*time.Hour)); !errors.Is(err, store.ErrShiftNotFinished) {
		t.Fatal("unfinished shift accepted", err)
	}
	exec("INSERT INTO calendar_days VALUES('2026-10-28',true,'future test')")
	exec(`INSERT INTO shift_instances(id,line_id,operational_date,code,starts_at,ends_at,required_staff,product)
 SELECT 'FUTURE-SHIFT',line_id,'2026-10-28','S1',$1,$2,required_staff,product FROM shift_instances WHERE id=$3`, domain.AsOf.Add(24*time.Hour), domain.AsOf.Add(32*time.Hour), seed.GoldenShift)
	exec("INSERT INTO production_lines SELECT 'FOREIGN',department_id,'Foreign line','OTHER-FACILITY' FROM production_lines WHERE id='A'")
	exec(`INSERT INTO shift_instances(id,line_id,operational_date,code,starts_at,ends_at,required_staff,product)
 SELECT 'FOREIGN-SHIFT','FOREIGN',operational_date,code,starts_at,ends_at,required_staff,product FROM shift_instances WHERE id=$1`, seed.GoldenShift)
	if _, err = service.Create(ctx, "FOREIGN-SHIFT", "operator"); !errors.Is(err, store.ErrNotFound) {
		t.Fatal("foreign facility exposed", err)
	}
	foreign := first
	foreign.ID = domain.ID()
	foreign.ShiftID = "FOREIGN-SHIFT"
	foreign.Snapshot.Shift.ID = "FOREIGN-SHIFT"
	foreign.Hash, err = domain.HandoverHash(foreign.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveHandover(ctx, foreign); !errors.Is(err, store.ErrNotFound) {
		t.Fatal("foreign facility saved", err)
	}
	// Overnight operational date follows the shift start, not its ending calendar date.
	night, err := service.Create(ctx, "A-20261026-S3", "operator")
	if err != nil || night.Snapshot.Shift.OperationalDate != "2026-10-26" || night.Snapshot.NextShift != nil {
		t.Fatal(night.Snapshot.Shift, err)
	}

	apiServer, err := httpapi.New(s, cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(apiServer)
	defer ts.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	request := func(method, path string, body []byte, key, origin string) (int, []byte) {
		t.Helper()
		req, err := http.NewRequest(method, ts.URL+"/api/v1"+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		raw, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		return res.StatusCode, raw
	}
	if status, _ := request("GET", "/shifts", nil, "", ""); status != 401 {
		t.Fatal(status)
	}
	if status, _ := request("GET", "/handovers/"+first.ID, nil, "", ""); status != 401 {
		t.Fatal(status)
	}
	if status, _ := request("POST", "/auth/login", store.JSON(map[string]string{"username": cfg.Username, "password": cfg.Password}), "", cfg.Origin); status != 200 {
		t.Fatal(status)
	}
	for _, path := range []string{"/shifts?limit=101", "/shifts?date=2026-02-30", "/handovers?offset=-1"} {
		if status, _ := request("GET", path, nil, "", ""); status != 400 {
			t.Fatal(path, status)
		}
	}
	status, raw := request("GET", "/shifts?line_id=A&date=2026-10-26&limit=2", nil, "", "")
	var catalogue struct {
		Items []domain.HandoverShift `json:"items"`
		Total int                    `json:"total"`
	}
	if err = json.Unmarshal(raw, &catalogue); err != nil || status != 200 || len(catalogue.Items) != 2 || catalogue.Total != 3 {
		t.Fatal(status, string(raw), err)
	}
	body := store.JSON(map[string]string{"shift_id": seed.GoldenShift})
	if status, _ = request("POST", "/handovers", body, "", cfg.Origin); status != 400 {
		t.Fatal("missing idempotency key", status)
	}
	if status, _ = request("POST", "/handovers", body, "foreign-origin", "https://example.invalid"); status != 403 {
		t.Fatal(status)
	}
	status, created := request("POST", "/handovers", body, "handover-create-1", cfg.Origin)
	if status != 201 {
		t.Fatal(status, string(created))
	}
	status, replay := request("POST", "/handovers", body, "handover-create-1", cfg.Origin)
	if status != 201 || !bytes.Equal(created, replay) {
		t.Fatal("idempotency replay changed")
	}
	if status, _ = request("POST", "/handovers", store.JSON(map[string]string{"shift_id": "A-20261026-S2"}), "handover-create-1", cfg.Origin); status != 409 {
		t.Fatal(status)
	}
	if status, _ = request("POST", "/handovers", []byte(`{"shift_id":"missing"}`), "handover-missing", cfg.Origin); status != 404 {
		t.Fatal(status)
	}
	if status, _ = request("POST", "/handovers", []byte(`{"shift_id":"FUTURE-SHIFT"}`), "handover-future", cfg.Origin); status != 409 {
		t.Fatal(status)
	}
	if status, _ = request("POST", "/handovers", []byte(`{"shift_id":"FOREIGN-SHIFT"}`), "handover-foreign", cfg.Origin); status != 404 {
		t.Fatal(status)
	}
	if status, _ = request("GET", "/handovers/"+first.ID, nil, "", ""); status != 200 {
		t.Fatal(status)
	}
	if status, _ = request("GET", "/handovers/missing", nil, "", ""); status != 404 {
		t.Fatal(status)
	}
	status, raw = request("GET", "/handovers?shift_id="+seed.GoldenShift+"&limit=1&offset=1", nil, "", "")
	var history struct {
		Items []domain.HandoverSummary `json:"items"`
		Total int                      `json:"total"`
	}
	if err = json.Unmarshal(raw, &history); err != nil || status != 200 || len(history.Items) != 1 || history.Total != 3 {
		t.Fatal(status, string(raw), err)
	}
	var runsAfter, budgetAfter, audits int
	if err = s.Pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM investigation_runs),(SELECT count(*) FROM budget_reservations),(SELECT count(*) FROM audit_events WHERE action='handover_created')").Scan(&runsAfter, &budgetAfter, &audits); err != nil {
		t.Fatal(err)
	}
	if runsBefore != runsAfter || budgetBefore != budgetAfter || audits != 4 {
		t.Fatal("handover changed model workload or lost audit", runsBefore, runsAfter, budgetBefore, budgetAfter, audits)
	}
}
