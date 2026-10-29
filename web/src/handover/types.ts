import type { Recommendation } from '../types'

export type Shift = {
  id: string
  line_id: string
  line_name: string
  operational_date: string
  code: string
  product: string
  starts_at: string
  ends_at: string
  working: boolean
}
export type Catalogue = {
  items: Shift[]
  total: number
  lines: { id: string; name: string }[]
  as_of: string
}
export type HandoverCase = {
  case_id: string
  rule_id: string
  severity: string
  status: string
  data_quality: string
  explanation: string
  run_id?: string
  run_status?: string
  model_mode?: string
  synthesis?: string
  recommendations: Recommendation[]
  evidence: { id: string; hash: string }[]
}
export type Handover = {
  id: string
  shift_id: string
  created_at: string
  created_by: string
  hash: string
  snapshot: {
    version: number
    facility_id: string
    timezone: string
    synthetic: boolean
    data_as_of: string
    shift: Shift
    next_shift: Shift | null
    sections: {
      key: string
      title: string
      quality: string
      facts: Record<string, string>
      notes: string[]
    }[]
    cases: HandoverCase[]
    downtime_events: {
      id: string
      machine_id: string
      machine_name: string
      starts_at: string
      ends_at: string
      reason: string
    }[]
    sources: { type: string; id: string }[]
    notes: string[]
  }
}
export type HandoverSummary = Pick<Handover, 'id' | 'shift_id' | 'created_at' | 'created_by'>
