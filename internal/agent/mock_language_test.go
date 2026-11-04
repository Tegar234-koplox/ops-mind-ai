package agent

import (
	"context"
	"encoding/json"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"strings"
	"testing"
)

func TestEnglishMockP1PreservesReferences(t *testing.T) {
	for _, rule := range []string{"ATT-001", "OT-002", "COST-001", "COST-002"} {
		t.Run(rule, func(t *testing.T) {
			messages := []Message{
				{Role: "user", Content: string(store.JSON(map[string]any{"shift_id": "fixture-shift", "case": domain.Case{RuleID: rule}}))},
				{Role: "assistant", ToolCalls: []Call{{Function: Function{Name: "get_evidence_record"}}}},
			}
			for _, name := range ToolNames {
				messages = append(messages, Message{Role: "tool", Content: string(store.JSON(domain.Evidence{ID: "evidence-" + name, Tool: name}))})
			}
			message, err := (MockModel{}).Complete(context.Background(), domain.Run{ReportLanguage: "en"}, messages)
			if err != nil {
				t.Fatal(err)
			}
			var report domain.Report
			if err = json.Unmarshal([]byte(message.Content), &report); err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(report.Summary, "SIMULATION: case "+rule) || len(report.Recommendations) != 1 || !strings.Contains(report.Recommendations[0].Title, "Review sources") {
				t.Fatal("English P1 prose", report)
			}
			if len(report.ObservationEvidenceIDs) != len(ToolNames) || len(report.Recommendations[0].EvidenceIDs) != len(ToolNames) {
				t.Fatal("Evidence references lost")
			}
		})
	}
}
