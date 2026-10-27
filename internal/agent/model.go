package agent

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"opsmind/internal/config"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"strconv"
	"strings"
	"time"
)

type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
type Call struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}
type Message struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ToolCalls        []Call `json:"tool_calls,omitempty"`
	ToolCallID       string `json:"tool_call_id,omitempty"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}
type Model interface {
	Complete(context.Context, domain.Run, []Message) (Message, error)
}

//go:embed report.schema.json
var ReportSchema []byte

type ProviderError struct {
	Status     int
	RetryAfter time.Duration
}

func (e ProviderError) Error() string   { return fmt.Sprintf("Meta API HTTP %d", e.Status) }
func (e ProviderError) Temporary() bool { return e.Status == 429 || e.Status >= 500 }

type MetaModel struct {
	Config config.Config
	Store  *store.Store
	Client *http.Client
}

func PriceMicros(input, output int64) int64 { return (input*125 + output*425 + 99) / 100 }
func (m *MetaModel) Complete(ctx context.Context, run domain.Run, messages []Message) (Message, error) {
	var schema any
	if e := json.Unmarshal(ReportSchema, &schema); e != nil {
		return Message{}, e
	}
	payload := map[string]any{"model": m.Config.Model, "messages": messages, "tools": Definitions(), "tool_choice": "auto", "parallel_tool_calls": true, "max_completion_tokens": 4096, "reasoning_effort": "low", "response_format": map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "investigation_report", "strict": true, "schema": schema}}}
	raw := store.JSON(payload)
	if len(raw) > 120000 {
		return Message{}, errors.New("konteks model melebihi batas payload")
	}
	// UTF-8 byte count is a conservative upper bound for visible input tokenization.
	maximum := PriceMicros(int64(len(raw))+1024, 4096)
	phase := m.Config.DevelopmentLimit
	if m.Config.Phase == "demo" {
		phase = m.Config.DemoLimit
	}
	reservation, e := m.Store.Reserve(ctx, run.ID, maximum, store.ReservationLimits{Total: m.Config.CreditLimit, Phase: phase, Run: m.Config.RunLimit, PhaseName: m.Config.Phase})
	if e != nil {
		return Message{}, e
	}
	req, e := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(m.Config.APIBase, "/")+"/chat/completions", bytes.NewReader(raw))
	if e != nil {
		return Message{}, e
	}
	req.Header.Set("Authorization", "Bearer "+m.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	client := m.Client
	if client == nil {
		client = &http.Client{Timeout: m.Config.ModelTimeout, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	}
	resp, e := client.Do(req)
	if e != nil {
		return Message{}, errors.New("Meta API tidak dapat dihubungi; reservasi konservatif dipertahankan")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if e != nil {
		return Message{}, errors.New("respons Meta tidak lengkap")
	}
	if resp.StatusCode != 200 {
		retry := time.Second
		if sec, x := strconv.Atoi(resp.Header.Get("Retry-After")); x == nil && sec > 0 {
			retry = time.Duration(sec) * time.Second
		} else if d, x := http.ParseTime(resp.Header.Get("Retry-After")); x == nil {
			retry = time.Until(d)
		}
		return Message{}, ProviderError{resp.StatusCode, retry}
	}
	var response struct {
		Choices []struct {
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			Input  int64 `json:"prompt_tokens"`
			Output int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if e = json.Unmarshal(body, &response); e != nil || len(response.Choices) != 1 {
		return Message{}, errors.New("format respons Meta tidak valid")
	}
	if response.Usage.Input <= 0 || response.Usage.Output < 0 {
		return Message{}, errors.New("usage Meta tidak tersedia; reservasi dipertahankan")
	}
	actual := PriceMicros(response.Usage.Input, response.Usage.Output)
	if e = m.Store.Settle(ctx, reservation, response.Usage.Input, response.Usage.Output, actual); e != nil {
		return Message{}, e
	}
	if actual > maximum {
		return Message{}, errors.New("usage melebihi reservasi; investigasi dihentikan")
	}
	if response.Choices[0].FinishReason == "length" {
		return Message{}, TruncatedResponse{}
	}
	if response.Choices[0].FinishReason == "content_filter" {
		return Message{}, errors.New("Meta menolak menghasilkan laporan")
	}
	return response.Choices[0].Message, nil
}

type MockModel struct{}

func (MockModel) Complete(ctx context.Context, run domain.Run, messages []Message) (Message, error) {
	select {
	case <-ctx.Done():
		return Message{}, ctx.Err()
	case <-time.After(180 * time.Millisecond):
	}
	evidence := []domain.Evidence{}
	for _, m := range messages {
		if m.Role == "tool" {
			var v domain.Evidence
			if json.Unmarshal([]byte(m.Content), &v) == nil && v.ID != "" {
				evidence = append(evidence, v)
			}
		}
	}
	shift := ""
	for _, m := range messages {
		if m.Role == "user" {
			var c struct {
				ShiftID string `json:"shift_id"`
			}
			if json.Unmarshal([]byte(m.Content), &c) == nil {
				shift = c.ShiftID
			}
		}
	}
	seen := map[string]bool{}
	byTool := map[string]string{}
	for _, e := range evidence {
		seen[e.Tool] = true
		byTool[e.Tool] = e.ID
	}
	groups := [][]string{{"get_overtime_variance", "get_cost_breakdown"}, {"get_shift_staffing", "get_attendance_summary", "get_approved_leave"}, {"get_production_performance", "get_downtime_events"}}
	for _, group := range groups {
		calls := []Call{}
		for _, name := range group {
			if !seen[name] {
				calls = append(calls, Call{ID: "mock-" + name, Type: "function", Function: Function{Name: name, Arguments: string(store.JSON(ToolArgs{ShiftID: shift}))}})
			}
		}
		if len(calls) > 0 {
			return Message{Role: "assistant", ToolCalls: calls}, nil
		}
	}
	looked := false
	for _, m := range messages {
		for _, c := range m.ToolCalls {
			if c.Function.Name == "get_evidence_record" {
				looked = true
			}
		}
	}
	if !looked {
		return Message{Role: "assistant", ToolCalls: []Call{{ID: "mock-open-evidence", Type: "function", Function: Function{Name: "get_evidence_record", Arguments: string(store.JSON(ToolArgs{ShiftID: shift, EvidenceID: byTool["get_downtime_events"]}))}}}}, nil
	}
	ids := []string{}
	for _, name := range ToolNames {
		if id := byTool[name]; id != "" {
			ids = append(ids, id)
		}
	}
	report := domain.Report{
		Summary:                "SIMULASI: bukti lintas departemen menunjukkan peningkatan lembur, kekurangan output, staffing yang perlu diverifikasi dan interupsi mesin. Hubungan ini belum membuktikan sebab-akibat.",
		ObservationEvidenceIDs: ids,
		Hypotheses:             []domain.Hypothesis{{Statement: "Downtime dan ketersediaan operator mungkin berkontribusi terhadap kebutuhan lembur dan output yang lebih rendah.", EvidenceIDs: []string{byTool["get_overtime_variance"], byTool["get_shift_staffing"], byTool["get_production_performance"], byTool["get_downtime_events"]}, Confidence: "medium", CounterEvidence: []string{"Waktu tunggu material dan perubahan product mix belum tersedia; dampak tiap faktor belum diukur."}}},
		InformationGaps:        []string{"Verifikasi catatan kehadiran yang belum lengkap; cuti disetujui bukan pelanggaran.", "Diperlukan urutan kejadian dan penilaian supervisor untuk menguji hubungan kausal."},
		Recommendations:        []domain.ProposedAction{{Title: "Verifikasi catatan kehadiran bersama supervisor", Rationale: "Lengkapi bukti sebelum menilai ketersediaan pekerja; kecualikan cuti yang sah.", EvidenceIDs: []string{byTool["get_attendance_summary"], byTool["get_approved_leave"]}, Department: "Workforce"}, {Title: "Tinjau interupsi mesin dan rencana pemulihan lini", Rationale: "Cocokkan log maintenance dan kehilangan output sebelum mengubah alokasi lembur.", EvidenceIDs: []string{byTool["get_downtime_events"], byTool["get_production_performance"]}, Department: "Maintenance"}}}
	return Message{Role: "assistant", Content: string(store.JSON(report))}, nil
}

type TruncatedResponse struct{}

func (TruncatedResponse) Error() string { return "Meta output terpotong oleh batas token" }
