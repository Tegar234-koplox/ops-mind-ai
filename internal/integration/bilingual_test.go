package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"opsmind/internal/agent"
	"opsmind/internal/config"
	"opsmind/internal/domain"
	"opsmind/internal/handover"
	"opsmind/internal/httpapi"
	"opsmind/internal/seed"
	"opsmind/internal/store"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBilingualInvestigationAndUpgrade(t *testing.T) {
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
	cfg := config.Config{Origin: "http://opsmind.test", Username: "operator", Password: "bilingual-test-password", SessionSecret: strings.Repeat("s", 40), Mode: "mock", MaxRounds: 8, MaxTools: 16, MaxRetries: 2, RunTimeout: 30 * time.Second, ModelTimeout: 5 * time.Second, OTThreshold: 2000, ProdThreshold: 1000, DowntimeThreshold: 60, AttendanceGrace: 15, CreditLimit: 20000000, DevelopmentLimit: 15000000, DemoLimit: 5000000, RunLimit: 250000, Phase: "development", EnableP1: true, CostThreshold: 2000}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server, err := httpapi.New(s, cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(server)
	defer ts.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	request := func(path string, body map[string]string, key string) (int, []byte) {
		t.Helper()
		req, err := http.NewRequest("POST", ts.URL+"/api/v1"+path, bytes.NewReader(store.JSON(body)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", cfg.Origin)
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
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
	if status, _ := request("/auth/login", map[string]string{"username": cfg.Username, "password": cfg.Password}, ""); status != 200 {
		t.Fatal("login", status)
	}
	for _, path := range []string{"/detections", "/cases/missing/investigations", "/cases/missing/follow-up-investigations"} {
		if status, _ := request(path, map[string]string{"report_language": "de", "instruction": "fixture"}, "invalid-language-"+path); status != 400 {
			t.Fatal("unsupported language", path, status)
		}
	}
	if status, raw := request("/detections", map[string]string{"shift_id": seed.GoldenShift, "report_language": "en"}, "english-detection"); status != 200 {
		t.Fatal(status, string(raw))
	}
	var caseID string
	if err = s.Pool.QueryRow(ctx, "SELECT id FROM anomaly_cases WHERE rule_id='OT-001' AND shift_id=$1", seed.GoldenShift).Scan(&caseID); err != nil {
		t.Fatal(err)
	}
	automatic, err := s.LatestRun(ctx, caseID)
	if err != nil || automatic.ReportLanguage != "en" {
		t.Fatal("automatic run language", automatic, err)
	}
	// Simulate an existing v5 snapshot whose case objects have no language field.
	data, err := s.HandoverData(ctx, seed.GoldenShift, domain.AsOf)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := handover.Build(data)
	for i := range snapshot.Cases {
		snapshot.Cases[i].ReportLanguage = ""
	}
	legacy := domain.Handover{ID: domain.ID(), ShiftID: seed.GoldenShift, CreatedBy: cfg.Username, CreatedAt: time.Now().UTC().Truncate(time.Microsecond), Snapshot: snapshot}
	legacy.Hash, err = domain.HandoverHash(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveHandover(ctx, legacy); err != nil {
		t.Fatal(err)
	}
	legacyJSON := store.JSON(legacy.Snapshot)
	down, err := os.ReadFile(filepath.Join("..", "store", "migrations", "000006_report_language.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	exec(string(down))
	exec("UPDATE schema_migrations SET version=5")
	s.Pool.Reset()
	if err = store.Migrate(dsn); err != nil {
		t.Fatal(err)
	}
	s.Pool.Reset()
	migrated, err := s.Run(ctx, automatic.ID)
	if err != nil || migrated.ReportLanguage != "id" {
		t.Fatal("legacy default", migrated, err)
	}
	after, err := s.Handover(ctx, legacy.ID)
	if err != nil || after.Hash != legacy.Hash || !bytes.Equal(store.JSON(after.Snapshot), legacyJSON) || !after.CreatedAt.Equal(legacy.CreatedAt) {
		t.Fatal("legacy immutable snapshot changed", err)
	}
	if err = s.Cancel(ctx, automatic.ID, cfg.Username); err != nil {
		t.Fatal(err)
	}
	path := "/cases/" + caseID + "/investigations"
	body := map[string]string{"report_language": "en"}
	status, raw := request(path, body, "english-investigation")
	if status != 202 {
		t.Fatal(status, string(raw))
	}
	var english domain.Run
	if err = json.Unmarshal(raw, &english); err != nil || english.ReportLanguage != "en" {
		t.Fatal(english, err)
	}
	if status, replayed := request(path, body, "english-investigation"); status != 202 || !bytes.Equal(raw, replayed) {
		t.Fatal("idempotency replay", status)
	}
	if status, _ := request(path, map[string]string{"report_language": "id"}, "english-investigation"); status != 409 {
		t.Fatal("language must be part of idempotency payload", status)
	}
	var inspected atomic.Bool
	model := simulatedModelFunc(func(callCtx context.Context, run domain.Run, messages []agent.Message) (agent.Message, error) {
		if run.ID == english.ID {
			if run.ReportLanguage != "en" || len(messages) == 0 || !strings.Contains(messages[0].Content, "Write all report prose in English") || strings.Contains(messages[0].Content, "Jawab bahasa Indonesia.") {
				return agent.Message{}, io.ErrUnexpectedEOF
			}
			inspected.Store(true)
		}
		return (agent.MockModel{}).Complete(callCtx, run, messages)
	})
	workerCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		(&agent.Runner{Store: s, Config: cfg, Model: model, Logger: logger}).Work(workerCtx)
	}()
	defer func() { stop(); <-done }()
	finished := awaitRun(t, s, english.ID)
	stop()
	<-done
	if finished.Status != "succeeded" || !inspected.Load() || finished.Report == nil || !strings.HasPrefix(finished.Report.Summary, "SIMULATION:") {
		t.Fatal("English simulation", finished.Status, finished.Error)
	}
	if finished.ReportLanguage != "en" || !strings.Contains(finished.Report.Recommendations[0].Title, "Verify attendance") {
		t.Fatal("English report fields")
	}
	reportBefore := store.JSON(finished.Report)
	current, err := (handover.Service{Store: s}).Create(ctx, seed.GoldenShift, cfg.Username)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range current.Snapshot.Cases {
		if item.RunID == english.ID {
			found = item.ReportLanguage == "en" && item.Synthesis == finished.Report.Summary
		}
	}
	if !found {
		t.Fatal("handover lost original report language")
	}
	status, raw = request("/cases/"+caseID+"/follow-up-investigations", map[string]string{"instruction": "Default language compatibility"}, "default-followup")
	if status != 202 {
		t.Fatal(status, string(raw))
	}
	var followup domain.Run
	if err = json.Unmarshal(raw, &followup); err != nil || followup.ReportLanguage != "id" {
		t.Fatal("default Indonesian", followup, err)
	}
	if err = s.Cancel(ctx, followup.ID, cfg.Username); err != nil {
		t.Fatal(err)
	}
	stored, err := s.Run(ctx, english.ID)
	if err != nil || !bytes.Equal(store.JSON(stored.Report), reportBefore) || stored.ReportLanguage != "en" {
		t.Fatal("stored English report changed", err)
	}
	old, err := s.Handover(ctx, legacy.ID)
	if err != nil || old.Hash != legacy.Hash || !bytes.Equal(store.JSON(old.Snapshot), legacyJSON) || !old.CreatedAt.Equal(legacy.CreatedAt) {
		t.Fatal("legacy snapshot changed after new reports", err)
	}
}
