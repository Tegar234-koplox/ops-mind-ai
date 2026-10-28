export type Case = {
  id: string
  rule_id: string
  entity_id: string
  shift_id: string
  department: string
  severity: string
  status: string
  data_quality: string
  expected: Record<string, string>
  observed: Record<string, string>
  explanation: string
  detected_at: string
}
export type Evidence = {
  id: string
  case_id: string
  run_id: string
  shift_id: string
  tool: string
  domain: string
  facts: Record<string, string>
  sources: { type: string; id: string }[]
  completeness: string
  notes: string[]
  hash: string
  retrieved_at: string
}
export type Action = {
  title: string
  rationale: string
  evidence_ids: string[]
  department: string
}
export type Report = {
  summary: string
  observation_evidence_ids: string[]
  hypotheses: {
    statement: string
    evidence_ids: string[]
    confidence: 'low' | 'medium'
    counter_evidence: string[]
  }[]
  information_gaps: string[]
  recommendations: Action[]
}
export type Run = {
  id: string
  case_id: string
  status: string
  mode: 'meta' | 'mock'
  instruction: string
  error?: string
  report?: Report
  model_rounds: number
  tool_calls: number
  cost_micros: number
  input_tokens: number
  output_tokens: number
  created_at: string
}
export type Recommendation = Action & {
  id: string
  run_id: string
  case_id: string
  decision?: string
  reason?: string
  reviewer?: string
}
export type InvestigationEvent = {
  id: number
  run_id: string
  kind: string
  message: string
  created_at: string
}
export type RunDetail = { run: Run; evidence: Evidence[]; events: InvestigationEvent[] }
export type CaseDetail = {
  case: Case
  latest_run: Run | null
  recommendations: Recommendation[]
  audit: {
    id: number
    actor: string
    action: string
    created_at: string
    details: Record<string, string>
  }[]
}
export type Overview = {
  active_rule_count: number
  cases: number
  active_runs: number
  pending_reviews: number
  employees: number
  facility: string
  model_mode: 'meta' | 'mock'
  model_name: string
  synthetic: boolean
  seed: number
  as_of: string
  budget_micros: Record<string, number>
  credit_limit_micros: number
  phase: string
}
