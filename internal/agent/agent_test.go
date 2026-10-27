package agent

import (
	"encoding/json"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"testing"
)

func TestToolScopeAndSchema(t *testing.T) {
	for _, tc := range []struct {
		name, args string
		ok         bool
	}{{"get_shift_staffing", `{"shift_id":"s1"}`, true}, {"get_shift_staffing", `{"shift_id":"other"}`, false}, {"run_sql", `{"shift_id":"s1"}`, false}, {"get_shift_staffing", `{"shift_id":"s1","sql":"select"}`, false}, {"get_evidence_record", `{"shift_id":"s1"}`, false}, {"get_shift_staffing", `{"shift_id":"s1"}{}`, false}} {
		_, e := ParseArgs(tc.name, tc.args, "s1")
		if (e == nil) != tc.ok {
			t.Fatalf("%+v: %v", tc, e)
		}
	}
}
func TestReportEvidenceIntegrity(t *testing.T) {
	ev := domain.Evidence{ID: "e1", ShiftID: "s1", Tool: "get_shift_staffing", Domain: "workforce", Facts: map[string]string{"verified_present": "31"}, Sources: []domain.SourceRef{}, Notes: []string{}, Completeness: "complete"}
	ev.Hash = EvidenceHash(ev)
	report := domain.Report{Summary: "Hipotesis sementara.", ObservationEvidenceIDs: []string{"e1"}, Hypotheses: []domain.Hypothesis{{Statement: "Staffing mungkin berkaitan.", EvidenceIDs: []string{"e1"}, Confidence: "low", CounterEvidence: []string{}}}, InformationGaps: []string{"Perlu verifikasi."}, Recommendations: []domain.ProposedAction{{Title: "Verifikasi", Rationale: "Butuh data tambahan.", EvidenceIDs: []string{"e1"}, Department: "Workforce"}}}
	if _, e := ValidateReport(store.JSON(report), []domain.Evidence{ev}); e != nil {
		t.Fatal(e)
	}
	report.Hypotheses[0].EvidenceIDs = []string{"foreign"}
	if _, e := ValidateReport(store.JSON(report), []domain.Evidence{ev}); e == nil {
		t.Fatal("foreign evidence accepted")
	}
	report.Hypotheses[0].EvidenceIDs = []string{"e1"}
	ev.Facts["verified_present"] = "99"
	if _, e := ValidateReport(store.JSON(report), []domain.Evidence{ev}); e == nil {
		t.Fatal("tampered snapshot accepted")
	}
	var invalid map[string]any
	_ = json.Unmarshal(store.JSON(report), &invalid)
	invalid["extra"] = "x"
	if _, e := ValidateReport(store.JSON(invalid), nil); e == nil {
		t.Fatal("additional field accepted")
	}
}
func TestCostCeiling(t *testing.T) {
	if PriceMicros(1000000, 1000000) != 5500000 {
		t.Fatal("token pricing incorrect")
	}
	if PriceMicros(1, 0) != 2 {
		t.Fatal("reservation must round up")
	}
}
