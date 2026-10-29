package integration

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"opsmind/internal/agent"
	"opsmind/internal/config"
	"opsmind/internal/detection"
	"opsmind/internal/domain"
	"opsmind/internal/seed"
	"opsmind/internal/store"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// All provider responses are local HTTP fixtures. No real Meta key or request is used.
func TestSimulatedMetaHTTPBoundary(t *testing.T) {
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
	if _, err = s.Pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); err != nil {
		t.Fatal(err)
	}
	if err = store.Migrate(dsn); err != nil {
		t.Fatal(err)
	}
	if _, err = seed.Run(ctx, s); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Mode: "mock", APIKey: "simulated-provider-key", Model: "simulated-meta",
		ModelTimeout: time.Second, Phase: "development",
		CreditLimit: 20000000, DevelopmentLimit: 15000000, DemoLimit: 5000000, RunLimit: 250000,
		OTThreshold: 2000, ProdThreshold: 1000, DowntimeThreshold: 60, AttendanceGrace: 15,
	}
	cases, err := (detection.Engine{Store: s, Config: cfg}).Execute(ctx, seed.GoldenShift, domain.AsOf)
	if err != nil {
		t.Fatal(err)
	}
	var primary domain.Case
	for _, c := range cases {
		if c.RuleID == "OT-001" {
			primary = c
		}
	}
	initial, err := s.LatestRun(ctx, primary.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Cancel(ctx, initial.ID, "simulated-provider-test"); err != nil {
		t.Fatal(err)
	}
	newRun := func(t *testing.T) domain.Run {
		t.Helper()
		run, err := s.Enqueue(ctx, primary.ID, "mock", "Simulasi batas HTTP, bukan Meta nyata", "test")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			current, err := s.Run(ctx, run.ID)
			if err != nil {
				t.Error(err)
				return
			}
			if !domain.Terminal(current.Status) {
				if err := s.Cancel(ctx, run.ID, "test"); err != nil {
					t.Error(err)
				}
			}
		})
		return run
	}
	// Leave a fresh, explicitly simulated fixture for standalone test + browser usage.
	t.Cleanup(func() {
		fresh, err := store.Open(ctx, dsn)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = fresh.Pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public")
		fresh.Pool.Close()
		if err != nil {
			t.Error(err)
			return
		}
		if err = store.Migrate(dsn); err != nil {
			t.Error(err)
			return
		}
		fresh, err = store.Open(ctx, dsn)
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
	message := []agent.Message{{Role: "user", Content: "Fixture HTTP sintetis tanpa data evaluator."}}
	response := func(finish string, input, output int64) string {
		return string(store.JSON(map[string]any{
			"choices": []any{map[string]any{"message": agent.Message{Role: "assistant", Content: "fixture-response"}, "finish_reason": finish}},
			"usage":   map[string]int64{"prompt_tokens": input, "completion_tokens": output},
		}))
	}
	tests := []struct {
		name      string
		status    int
		body      string
		retry     string
		temporary bool
		settled   bool
		truncated bool
		valid     bool
	}{
		{name: "success_usage_settled", status: 200, body: response("stop", 100, 20), settled: true, valid: true},
		{name: "rate_limit", status: 429, body: `{"error":"sensitive-provider-marker"}`, retry: "2", temporary: true},
		{name: "server_unavailable", status: 503, body: `{"error":"sensitive-provider-marker"}`, temporary: true},
		{name: "authentication_not_retryable", status: 401, body: `{"error":"sensitive-provider-marker"}`},
		{name: "malformed_envelope", status: 200, body: `{"choices":`},
		{name: "missing_usage", status: 200, body: `{"choices":[{"message":{"role":"assistant","content":"x"}}]}`},
		{name: "negative_usage", status: 200, body: response("stop", 100, -1)},
		{name: "no_choice", status: 200, body: `{"choices":[],"usage":{"prompt_tokens":100,"completion_tokens":20}}`},
		{name: "truncated_response_is_billed", status: 200, body: response("length", 100, 20), settled: true, truncated: true},
		{name: "content_filter_is_billed", status: 200, body: response("content_filter", 100, 20), settled: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := newRun(t)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.URL.Path != "/chat/completions" || r.Method != http.MethodPost {
					t.Error("unexpected HTTP route")
				}
				if r.Header.Get("Authorization") != "Bearer "+cfg.APIKey {
					t.Error("provider auth header missing")
				}
				var payload map[string]any
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				if payload["model"] != cfg.Model || payload["response_format"] == nil || payload["tools"] == nil {
					t.Error("model contract missing")
				}
				if tc.retry != "" {
					w.Header().Set("Retry-After", tc.retry)
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			local := cfg
			local.APIBase = server.URL
			model := agent.MetaModel{Config: local, Store: s}
			msg, err := model.Complete(ctx, run, message)
			if (err == nil) != tc.valid {
				t.Fatalf("unexpected result: %v", err)
			}
			if tc.valid && msg.Content != "fixture-response" {
				t.Fatal("response lost")
			}
			if err != nil && (strings.Contains(err.Error(), cfg.APIKey) || strings.Contains(err.Error(), "sensitive-provider-marker")) {
				t.Fatal("provider details leaked")
			}
			if tc.status != 200 {
				var pe agent.ProviderError
				if !errors.As(err, &pe) || pe.Status != tc.status || pe.Temporary() != tc.temporary {
					t.Fatal("provider classification", err)
				}
				if tc.retry != "" && pe.RetryAfter != 2*time.Second {
					t.Fatal("Retry-After lost")
				}
			}
			if tc.truncated {
				var truncated agent.TruncatedResponse
				if !errors.As(err, &truncated) {
					t.Fatal("truncation not classified")
				}
			}
			if calls.Load() != 1 {
				t.Fatal("adapter retried without orchestrator")
			}
			var count, settled int
			var actual int64
			if err = s.Pool.QueryRow(ctx, "SELECT count(*),count(actual_micros),coalesce(sum(actual_micros),0) FROM budget_reservations WHERE run_id=$1", run.ID).Scan(&count, &settled, &actual); err != nil {
				t.Fatal(err)
			}
			if count != 1 || (settled == 1) != tc.settled {
				t.Fatal("reservation reconciliation", count, settled)
			}
			if tc.settled && actual != agent.PriceMicros(100, 20) {
				t.Fatal("usage cost mismatch", actual)
			}
		})
	}
	t.Run("redirect_does_not_forward_key", func(t *testing.T) {
		run := newRun(t)
		var leaked atomic.Int32
		sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Add(1); w.WriteHeader(200) }))
		defer sink.Close()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, sink.URL, http.StatusFound) }))
		defer server.Close()
		local := cfg
		local.APIBase = server.URL
		_, err := (&agent.MetaModel{Config: local, Store: s}).Complete(ctx, run, message)
		var pe agent.ProviderError
		if !errors.As(err, &pe) || pe.Status != 302 || leaked.Load() != 0 {
			t.Fatal("redirect followed", err)
		}
	})
	t.Run("timeout_keeps_conservative_reservation", func(t *testing.T) {
		run := newRun(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
			}
		}))
		defer server.Close()
		local := cfg
		local.APIBase = server.URL
		local.ModelTimeout = 30 * time.Millisecond
		_, err := (&agent.MetaModel{Config: local, Store: s}).Complete(ctx, run, message)
		if err == nil || !strings.Contains(err.Error(), "reservasi konservatif") {
			t.Fatal("timeout lost reservation policy", err)
		}
		var unresolved int
		if err = s.Pool.QueryRow(ctx, "SELECT count(*) FROM budget_reservations WHERE run_id=$1 AND actual_micros IS NULL", run.ID).Scan(&unresolved); err != nil || unresolved != 1 {
			t.Fatal(unresolved, err)
		}
	})
	t.Run("budget_and_payload_stop_before_http", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
		defer server.Close()
		for _, oversized := range []bool{false, true} {
			run := newRun(t)
			local := cfg
			local.APIBase = server.URL
			messages := message
			if oversized {
				messages = []agent.Message{{Role: "user", Content: strings.Repeat("x", 120001)}}
			} else {
				local.RunLimit = 1
			}
			_, err := (&agent.MetaModel{Config: local, Store: s}).Complete(ctx, run, messages)
			if err == nil {
				t.Fatal("request outside limits accepted")
			}
			var reserved int
			if err = s.Pool.QueryRow(ctx, "SELECT count(*) FROM budget_reservations WHERE run_id=$1", run.ID).Scan(&reserved); err != nil || reserved != 0 {
				t.Fatal(reserved, err)
			}
			if err = s.Cancel(ctx, run.ID, "test"); err != nil {
				t.Fatal(err)
			}
		}
		if calls.Load() != 0 {
			t.Fatal("network contacted outside budget or payload limit")
		}
	})
	t.Run("orchestrator_failure_limits", func(t *testing.T) {
		for _, scenario := range []struct {
			name         string
			calls, tools int
			status       string
		}{
			{"retry_429", 3, 0, "failed"},
			{"invalid_json_repaired_once", 2, 0, "failed"},
			{"truncation_repaired_once", 2, 0, "failed"},
			{"unknown_tool", 1, 0, "failed"},
			{"foreign_shift", 1, 0, "failed"},
			{"model_deadline", 1, 0, "failed"},
			{"incomplete_domains", 2, 1, "insufficient_evidence"},
		} {
			t.Run(scenario.name, func(t *testing.T) {
				run := newRun(t)
				var calls atomic.Int32
				model := simulatedModelFunc(func(callCtx context.Context, r domain.Run, messages []agent.Message) (agent.Message, error) {
					calls.Add(1)
					switch scenario.name {
					case "retry_429":
						return agent.Message{}, agent.ProviderError{Status: 429}
					case "truncation_repaired_once":
						return agent.Message{}, agent.TruncatedResponse{}
					case "model_deadline":
						<-callCtx.Done()
						return agent.Message{}, callCtx.Err()
					case "unknown_tool":
						return agent.Message{Role: "assistant", ToolCalls: []agent.Call{{ID: "fixture-call", Type: "function", Function: agent.Function{Name: "run_sql", Arguments: `{}`}}}}, nil
					case "foreign_shift":
						return agent.Message{Role: "assistant", ToolCalls: []agent.Call{{ID: "fixture-call", Type: "function", Function: agent.Function{Name: "get_shift_staffing", Arguments: `{"shift_id":"foreign"}`}}}}, nil
					case "incomplete_domains":
						for _, message := range messages {
							if message.Role == "tool" {
								var evidence domain.Evidence
								if err := json.Unmarshal([]byte(message.Content), &evidence); err != nil {
									return agent.Message{}, err
								}
								ids := []string{evidence.ID}
								report := domain.Report{Summary: "SIMULASI: bukti belum lengkap.", ObservationEvidenceIDs: ids, Hypotheses: []domain.Hypothesis{{Statement: "Ketersediaan operator perlu diverifikasi.", EvidenceIDs: ids, Confidence: "low", CounterEvidence: []string{}}}, InformationGaps: []string{"Domain lain belum diambil."}, Recommendations: []domain.ProposedAction{{Title: "Verifikasi", Rationale: "Butuh bukti tambahan.", EvidenceIDs: ids, Department: "Workforce"}}}
								return agent.Message{Role: "assistant", Content: string(store.JSON(report))}, nil
							}
						}
						return agent.Message{Role: "assistant", ToolCalls: []agent.Call{{ID: "fixture-call", Type: "function", Function: agent.Function{Name: "get_shift_staffing", Arguments: string(store.JSON(agent.ToolArgs{ShiftID: seed.GoldenShift}))}}}}, nil
					default:
						return agent.Message{Role: "assistant", Content: "{"}, nil
					}
				})
				local := cfg
				local.RunTimeout = 15 * time.Second
				local.ModelTimeout = 30 * time.Millisecond
				local.MaxRetries = 2
				local.MaxRounds = 8
				local.MaxTools = 16
				workerCtx, stop := context.WithCancel(ctx)
				done := make(chan struct{})
				go func() {
					defer close(done)
					(&agent.Runner{Store: s, Config: local, Model: model, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Work(workerCtx)
				}()
				defer func() { stop(); <-done }()
				finished := awaitRun(t, s, run.ID)
				if finished.Status != scenario.status || int(calls.Load()) != scenario.calls || finished.ToolCalls != scenario.tools {
					t.Fatal("orchestrator bounds", finished.Status, calls.Load(), finished.ToolCalls)
				}
				if scenario.status == "insufficient_evidence" && (finished.Report == nil || len(finished.Report.Recommendations) != 0) {
					t.Fatal("incomplete evidence produced actionable recommendations")
				}
			})
		}
	})
}

type simulatedModelFunc func(context.Context, domain.Run, []agent.Message) (agent.Message, error)

func (f simulatedModelFunc) Complete(ctx context.Context, run domain.Run, messages []agent.Message) (agent.Message, error) {
	return f(ctx, run, messages)
}
