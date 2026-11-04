package agent

import "opsmind/internal/domain"

// English mock prose follows the same evidence IDs and limits as the Indonesian fixture.
func englishMockReport(report *domain.Report, rule string, ids []string) {
	if rule == "ATT-001" || rule == "OT-002" || rule == "COST-001" || rule == "COST-002" {
		report.Summary = "SIMULATION: case " + rule + " was examined using source records and cross-department evidence. Human review is required; no violation or efficiency is established."
		report.Hypotheses = []domain.Hypothesis{{Statement: "Reconcile the finding with source completeness and operational context before deciding on action.", EvidenceIDs: ids, Confidence: "low", CounterEvidence: []string{"Available evidence does not establish causation or worker misconduct."}}}
		report.InformationGaps = []string{"A supervisor explanation and verification of operational context are needed."}
		report.Recommendations = []domain.ProposedAction{{Title: "Review sources and confirm the context of case " + rule, Rationale: "Use stored evidence before making operational decisions; this review does not change business data.", EvidenceIDs: ids, Department: "Operations"}}
		return
	}
	report.Summary = "SIMULATION: cross-department evidence shows increased overtime, lower output, staffing records requiring verification and machine interruptions. These correlations do not establish causation."
	report.Hypotheses[0].Statement = "Downtime and operator availability may contribute to overtime requirements and lower output."
	report.Hypotheses[0].CounterEvidence = []string{"Material waiting times and product-mix changes are unavailable; individual effects have not been measured."}
	report.InformationGaps = []string{"Verify incomplete attendance records; approved leave is legitimate.", "A timeline and supervisor assessment are needed to test causal relationships."}
	report.Recommendations[0].Title = "Verify attendance records with the supervisor"
	report.Recommendations[0].Rationale = "Complete the evidence before assessing worker availability; exclude approved leave."
	report.Recommendations[1].Title = "Review machine interruptions and the line recovery plan"
	report.Recommendations[1].Rationale = "Reconcile maintenance logs and output losses before changing overtime allocation."

}
