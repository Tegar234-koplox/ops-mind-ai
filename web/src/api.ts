export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}
export async function api<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch('/api/v1' + path, {
    credentials: 'same-origin',
    ...(body === undefined
      ? {}
      : {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Idempotency-Key': crypto.randomUUID() },
          body: JSON.stringify(body),
        }),
  })
  const data = await res.json().catch(() => null)
  if (!res.ok)
    throw new ApiError(
      data?.error?.message ?? 'Layanan belum siap. Coba lagi setelah beberapa saat.',
      res.status,
    )
  return data as T
}
export const active = (s?: string) => s === 'queued' || s === 'running'
export const stateLabel: Record<string, string> = {
  open: 'Terbuka',
  investigating: 'Diinvestigasi',
  review_required: 'Perlu review',
  reviewed: 'Sudah direview',
  queued: 'Dalam antrean',
  running: 'Sedang berjalan',
  succeeded: 'Selesai',
  insufficient_evidence: 'Bukti belum cukup',
  failed: 'Gagal',
  cancelled: 'Dibatalkan',
  confirmed: 'Anomali terkonfirmasi',
  needs_verification: 'Perlu verifikasi',
  complete: 'Lengkap',
  approved: 'Disetujui',
  rejected: 'Ditolak',
}
export const ruleLabel: Record<string, string> = {
  'OT-001': 'Biaya lembur meningkat',
  'PROD-001': 'Output di bawah target',
  'ATT-002': 'Kehadiran perlu diverifikasi',
  'MAINT-001': 'Downtime melebihi batas',
  'ATT-001': 'Rangkaian absen terkonfirmasi',
  'OT-002': 'Lembur tanpa persetujuan berlaku',
  'COST-001': 'Pengeluaran di atas baseline',
  'COST-002': 'Pengeluaran di bawah baseline',
}
export const toolLabel: Record<string, string> = {
  get_shift_staffing: 'Ketersediaan operator',
  get_attendance_summary: 'Rekonsiliasi kehadiran',
  get_approved_leave: 'Cuti yang disetujui',
  get_overtime_variance: 'Variansi biaya lembur',
  get_cost_breakdown: 'Rincian pengeluaran',
  get_production_performance: 'Kinerja produksi',
  get_downtime_events: 'Interupsi mesin',
  get_evidence_record: 'Validasi sumber bukti',
}
export const factLabel: Record<string, string> = {
  confirmed_absent: 'Absen terkonfirmasi',
  confirmed_absence_workers: 'Pekerja dengan rangkaian absen',
  consecutive_shifts: 'Shift berturut-turut',
  required_consecutive_shifts: 'Ambang rangkaian shift',
  employee_ids: 'ID pekerja sintetis',
  unverified_unapproved_records: 'Persetujuan lembur perlu verifikasi',
  unapproved_records: 'Catatan tanpa persetujuan',
  unapproved_minutes: 'Durasi tanpa persetujuan',
  unverified_overtime_records: 'Lembur belum terverifikasi',
  expense_baseline_rupiah: 'Baseline pengeluaran',
  expense_actual_rupiah: 'Pengeluaran aktual',
  expense_variance_bps: 'Selisih pengeluaran',
  expense_posting_closed: 'Posting pengeluaran ditutup',
  scheduled: 'Operator terjadwal',
  verified_present: 'Hadir terverifikasi',
  approved_leave: 'Cuti disetujui',
  unverified: 'Belum terverifikasi',
  required_staff: 'Kebutuhan operator',
  currency: 'Mata uang',
  observed_rupiah: 'Lembur aktual',
  baseline_rupiah: 'Baseline lembur',
  variance_bps: 'Kenaikan biaya',
  comparable_shifts: 'Shift pembanding',
  baseline_basis: 'Dasar pembanding',
  overtime_rupiah: 'Biaya lembur',
  maintenance_rupiah: 'Biaya maintenance',
  planned_output: 'Target produksi',
  actual_output: 'Output aktual',
  shortfall_bps: 'Kekurangan output',
  line_downtime_minutes: 'Downtime lini',
  aggregation: 'Metode agregasi',
  output: 'Output',
  downtime_minutes: 'Downtime',
  threshold_minutes: 'Ambang menit',
  threshold_bps: 'Ambang variansi',
  complete_required_punches: 'Kehadiran wajib punch',
  minimum_comparable_shifts: 'Minimum pembanding',
}
export function factValue(k: string, v: string) {
  if (k === 'expense_posting_closed') return v === 'true' ? 'Ya' : 'Belum'
  if (k.endsWith('_rupiah'))
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      maximumFractionDigits: 0,
    }).format(BigInt(v))
  if (k.endsWith('_bps')) return (Number(v) / 100).toLocaleString('id-ID') + '%'
  if (k.endsWith('_minutes')) return v + ' menit'
  return v
}
export const timestamp = (v: string) =>
  new Intl.DateTimeFormat('id-ID', {
    timeZone: 'Asia/Bangkok',
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(v)) + ' WIB'
