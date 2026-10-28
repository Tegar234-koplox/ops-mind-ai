import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Activity,
  ArrowRight,
  Bell,
  Check,
  ChevronRight,
  CircleDot,
  Factory,
  Fingerprint,
  Layers3,
  LayoutList,
  LogOut,
  Play,
  Search,
  ShieldCheck,
  Sparkles,
  Square,
  Workflow,
} from 'lucide-react'
import {
  active,
  api,
  ApiError,
  factLabel,
  factValue,
  ruleLabel,
  stateLabel,
  timestamp,
  toolLabel,
} from './api'
import { Badge, Busy, Empty, ErrorMessage, EvidenceCard, EvidenceDialog } from './components'
import type { Case, CaseDetail, Evidence, Overview, Run, RunDetail } from './types'
type Session = { username: string; model_mode: string }
export default function App() {
  const client = useQueryClient()
  const session = useQuery({
    queryKey: ['session'],
    queryFn: async () => {
      try {
        return await api<Session>('/auth/session')
      } catch (e) {
        if (e instanceof ApiError && e.status === 401) return null
        throw e
      }
    },
    retry: false,
  })
  if (session.isPending)
    return (
      <div className="app-loading">
        <Brand />
        <Busy text="Menghubungkan ruang investigasi…" />
      </div>
    )
  if (session.error)
    return (
      <div className="app-loading">
        <Brand />
        <ErrorMessage message={session.error.message} />
        <button className="primary" onClick={() => session.refetch()}>
          Coba lagi
        </button>
      </div>
    )
  if (!session.data)
    return <Login onSuccess={() => client.invalidateQueries({ queryKey: ['session'] })} />
  return <Workspace session={session.data} />
}
function Brand() {
  return (
    <div className="brand">
      <span className="brand-mark">
        <Workflow size={23} />
      </span>
      <span>
        OpsMind<span className="brand-ai">AI</span>
      </span>
    </div>
  )
}
function Login({ onSuccess }: { onSuccess: () => void }) {
  const [username, setUsername] = useState('operator'),
    [password, setPassword] = useState('')
  const login = useMutation({
    mutationFn: () => api('/auth/login', { username, password }),
    onSuccess,
  })
  return (
    <main className="login-page">
      <section className="login-story">
        <Brand />
        <div className="login-copy">
          <span className="eyebrow light">Manufacturing operations intelligence</span>
          <h1>
            Setiap anomali
            <br />
            punya cerita.
            <br />
            <em>Temukan buktinya.</em>
          </h1>
          <p>
            Investigasi lintas departemen yang menghubungkan data operasional dengan keputusan
            manusia.
          </p>
          <div className="login-flow">
            <span>
              <Activity size={18} /> Deteksi
            </span>
            <ChevronRight size={14} />
            <span>
              <Search size={18} /> Investigasi
            </span>
            <ChevronRight size={14} />
            <span>
              <ShieldCheck size={18} /> Review
            </span>
          </div>
        </div>
        <div className="login-foot">
          <Factory size={16} /> Satu fasilitas. Bukti dari seluruh operasi.
        </div>
      </section>
      <section className="login-form-wrap">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            login.mutate()
          }}
          className="login-form"
        >
          <span className="login-lock">
            <Fingerprint size={30} />
          </span>
          <span className="eyebrow">Akses operator</span>
          <h2>Masuk ke ruang investigasi</h2>
          <p className="muted">Gunakan kredensial operator pada konfigurasi aplikasi.</p>
          <label htmlFor="username">Nama pengguna</label>
          <input
            id="username"
            autoComplete="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <label htmlFor="password">Kata sandi</label>
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          {login.error ? <ErrorMessage message={login.error.message} /> : null}
          <button className="primary" disabled={login.isPending}>
            {login.isPending ? 'Memeriksa akses…' : 'Masuk'}
            <ArrowRight size={17} />
          </button>
          <p className="login-disclaimer">
            <ShieldCheck size={15} /> Data manufaktur sintetis. Keputusan operasional tetap
            memerlukan peninjau manusia.
          </p>
        </form>
      </section>
    </main>
  )
}
function Workspace({ session }: { session: Session }) {
  const client = useQueryClient(),
    [selected, setSelected] = useState(''),
    [search, setSearch] = useState(''),
    [tab, setTab] = useState('live'),
    [openEvidence, setOpenEvidence] = useState<Evidence | null>(null),
    [error, setError] = useState('')
  const overview = useQuery({
    queryKey: ['overview'],
    queryFn: () => api<Overview>('/overview'),
    refetchInterval: (q) => (q.state.data?.active_runs ? 3000 : false),
  })
  const cases = useQuery({
    queryKey: ['cases'],
    queryFn: () => api<{ items: Case[]; total: number }>('/cases?limit=100'),
    refetchInterval: (q) =>
      q.state.data?.items.some((c) => c.status === 'investigating') ? 3000 : false,
  })
  const items = cases.data?.items ?? [],
    caseID = selected || items[0]?.id
  const detail = useQuery({
    queryKey: ['case', caseID],
    queryFn: () => api<CaseDetail>('/cases/' + caseID),
    enabled: !!caseID,
    refetchInterval: (q) => (active(q.state.data?.latest_run?.status) ? 3000 : false),
  })
  const runID = detail.data?.latest_run?.id
  const investigation = useQuery({
    queryKey: ['run', runID],
    queryFn: () => api<RunDetail>('/investigations/' + runID),
    enabled: !!runID,
    refetchInterval: (q) => (active(q.state.data?.run.status) ? 3000 : false),
  })
  const invalidate = () => {
    for (const key of ['overview', 'cases', 'case', 'run'])
      void client.invalidateQueries({ queryKey: [key] })
  }
  useEffect(() => {
    if (!runID || !active(detail.data?.latest_run?.status)) return
    const stream = new EventSource('/api/v1/investigations/' + runID + '/events')
    stream.addEventListener('progress', () => {
      for (const key of [['run', runID], ['case', caseID], ['overview']])
        void client.invalidateQueries({ queryKey: key })
    })
    stream.onerror = () => stream.close()
    return () => stream.close()
  }, [runID, detail.data?.latest_run?.status, caseID, client])
  const detect = useMutation({
    mutationFn: () => api('/detections', {}),
    onSuccess: () => {
      setError('')
      invalidate()
    },
    onError: (e) => setError(e.message),
  })
  const start = useMutation({
    mutationFn: (instruction: string) =>
      api<Run>(
        '/cases/' + caseID + (instruction ? '/follow-up-investigations' : '/investigations'),
        { instruction },
      ),
    onSuccess: async (run) => {
      const caseKey = ['case', run.case_id]
      await client.cancelQueries({ queryKey: caseKey })
      client.setQueryData<CaseDetail>(caseKey, (previous) =>
        previous ? { ...previous, latest_run: run, recommendations: [] } : previous,
      )
      client.setQueryData<RunDetail>(['run', run.id], { run, evidence: [], events: [] })
      if (caseID === run.case_id) setTab('live')
      setError('')
      invalidate()
    },
    onError: (e) => setError(e.message),
  })
  const cancel = useMutation({
    mutationFn: () => api('/investigations/' + runID + '/cancel', {}),
    onSuccess: invalidate,
    onError: (e) => setError(e.message),
  })
  const logout = useMutation({
    mutationFn: () => api('/auth/logout', {}),
    onSuccess: async () => {
      await client.cancelQueries()
      client.setQueryData(['session'], null)
      client.removeQueries({ predicate: (q) => q.queryKey[0] !== 'session' })
    },
    onError: (e) => setError(e.message),
  })
  const data = overview.data,
    visible = items.filter((c) =>
      (ruleLabel[c.rule_id] + ' ' + c.department + ' ' + c.rule_id)
        .toLowerCase()
        .includes(search.toLowerCase()),
    )
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <Brand />
        <div className="workspace-label">WORKSPACE</div>
        <nav aria-label="Navigasi utama">
          <button className="nav-item selected" onClick={() => setTab('live')}>
            <LayoutList size={18} /> Investigasi <span>{data?.cases ?? 0}</span>
          </button>
          <button className="nav-item" onClick={() => setTab('report')}>
            <Layers3 size={18} /> Laporan bukti
          </button>
          <button className="nav-item" onClick={() => setTab('review')}>
            <ShieldCheck size={18} /> Review manusia <i className="nav-dot" />
          </button>
        </nav>
        <div className="sidebar-bottom">
          <div className="facility">
            <span>
              <Factory size={19} />
            </span>
            <div>
              <strong>Fasilitas demo</strong>
              <small>FACTORY-01 · Asia/Bangkok</small>
            </div>
          </div>
          <div className="sidebar-note">
            <span className="status-dot" /> Data sintetis · Seed 20261027
          </div>
        </div>
      </aside>
      <div className="workspace">
        <header className="topbar">
          <div className="breadcrumb">
            Operasional <ChevronRight size={13} />
            <strong>Investigasi</strong>
          </div>
          <div className="operator">
            <span className="top-status">
              <span className="status-dot" /> Workspace terlindungi
            </span>
            <span className="avatar">{session.username[0].toUpperCase()}</span>
            <span>{session.username}</span>
            <button className="icon-button" aria-label="Keluar" onClick={() => logout.mutate()}>
              <LogOut size={17} />
            </button>
          </div>
        </header>
        <main className="main-content">
          <div className="page-heading">
            <div>
              <span className="eyebrow">OPERATIONS INTELLIGENCE</span>
              <h1>
                Ruang investigasi<span>.</span>
              </h1>
              <p>Hubungkan anomali, telusuri bukti, ambil keputusan.</p>
            </div>
            <button className="primary" onClick={() => detect.mutate()} disabled={detect.isPending}>
              <Activity size={17} />
              {detect.isPending ? 'Mendeteksi…' : 'Jalankan deteksi'}
            </button>
          </div>
          <div className={'mode-banner ' + (data?.model_mode === 'mock' ? 'mock' : 'meta')}>
            <Sparkles size={17} />
            <strong>{data?.model_mode === 'mock' ? 'SIMULASI MODEL' : 'META AI'}</strong>
            <span>
              {data?.model_mode === 'mock'
                ? 'Respons model tiruan untuk pengujian. Data, tool, dan penyimpanan tetap nyata.'
                : (data?.model_name ?? 'Menghubungkan model') +
                  ' · Agent memilih bukti melalui tool baca terbatas.'}
            </span>
          </div>
          {error ? <ErrorMessage message={error} /> : null}
          {overview.error ? <ErrorMessage message={overview.error.message} /> : null}
          <section className="stat-grid" aria-label="Status workspace">
            <Stat
              label="Kasus terdeteksi"
              value={data?.cases}
              icon={<LayoutList />}
              note={(data?.active_rule_count ?? 4) + ' aturan aktif'}
            />
            <Stat
              label="Investigasi aktif"
              value={data?.active_runs}
              icon={<Activity />}
              note="Proses dengan checkpoint"
            />
            <Stat
              label="Menunggu review"
              value={data?.pending_reviews}
              icon={<Bell />}
              note="Keputusan ada pada manusia"
            />
            <Stat
              label="Terpakai / direservasi"
              value={data ? 'US$' + ((data.budget_micros.total ?? 0) / 1e6).toFixed(3) : undefined}
              icon={<CircleDot />}
              note={'Batas aplikasi US$' + (data?.credit_limit_micros ?? 20000000) / 1e6}
            />
          </section>
          <section className="investigation-workspace">
            <div className="inbox">
              <div className="panel-heading">
                <h2>
                  Inbox kasus <span>{items.length}</span>
                </h2>
                <span className="subtle-tag">SEMUA LINI</span>
              </div>
              <label className="search">
                <Search size={16} />
                <input
                  aria-label="Cari kasus"
                  placeholder="Cari kasus atau departemen…"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </label>
              <div className="case-list">
                {cases.isPending ? (
                  <Busy />
                ) : cases.error ? (
                  <ErrorMessage message={cases.error.message} />
                ) : visible.length ? (
                  visible.map((c) => (
                    <button
                      key={c.id}
                      className={'case-item ' + (c.id === caseID ? 'active' : '')}
                      onClick={() => {
                        setSelected(c.id)
                        setTab('live')
                      }}
                    >
                      <div className="case-item-top">
                        <span className={'case-rule ' + c.severity}>{c.rule_id}</span>
                        <span className="case-department">{c.department}</span>
                      </div>
                      <h3>{ruleLabel[c.rule_id] ?? c.rule_id}</h3>
                      <p>Assembly Line {c.entity_id} · Shift operasional</p>
                      <div className="case-item-bottom">
                        <Badge value={c.severity} />
                        <span>{stateLabel[c.status] ?? c.status}</span>
                      </div>
                    </button>
                  ))
                ) : (
                  <Empty title={search ? 'Kasus tidak ditemukan' : 'Belum ada kasus'}>
                    <p>
                      {search
                        ? 'Gunakan kata kunci lain.'
                        : 'Jalankan deteksi untuk memeriksa lini produksi.'}
                    </p>
                  </Empty>
                )}
              </div>
              <div className="inbox-foot">
                <ShieldCheck size={13} /> Seluruh catatan adalah data sintetis
              </div>
            </div>
            <div className="case-panel">
              {!caseID ? (
                <Empty title="Mulai dari satu pertanyaan">
                  <p>
                    Mengapa biaya lembur meningkat saat output menurun?
                    <br />
                    Jalankan deteksi untuk menghubungkan buktinya.
                  </p>
                  <button
                    className="secondary"
                    disabled={detect.isPending}
                    onClick={() => detect.mutate()}
                  >
                    Deteksi anomali <ArrowRight size={15} />
                  </button>
                </Empty>
              ) : detail.isPending ? (
                <Busy text="Memuat konteks kasus…" />
              ) : detail.error ? (
                <ErrorMessage message={detail.error.message} />
              ) : detail.data ? (
                <>
                  <div className="case-header">
                    <div>
                      <span className="eyebrow">
                        ASSEMBLY LINE {detail.data.case.entity_id} <span> / </span>{' '}
                        {detail.data.case.rule_id}
                      </span>
                      <h2>{ruleLabel[detail.data.case.rule_id] ?? detail.data.case.rule_id}</h2>
                      <p>{detail.data.case.explanation}</p>
                    </div>
                    <Badge value={detail.data.case.data_quality} />
                  </div>
                  <div className="case-numbers">
                    {Object.entries(detail.data.case.observed).map(([k, v]) => (
                      <div key={k}>
                        <span>{factLabel[k] ?? k}</span>
                        <strong>{factValue(k, v)}</strong>
                      </div>
                    ))}
                  </div>
                  <div className="tabs" role="tablist" aria-label="Tahap investigasi">
                    {[
                      ['live', 'Investigasi langsung'],
                      ['report', 'Laporan bukti'],
                      ['review', 'Review manusia'],
                    ].map(([id, label], i) => (
                      <button
                        role="tab"
                        id={'tab-' + id}
                        aria-controls="investigation-panel"
                        aria-selected={tab === id}
                        key={id}
                        className={tab === id ? 'selected' : ''}
                        onClick={() => setTab(id)}
                      >
                        <span>0{i + 1}</span>
                        {label}
                      </button>
                    ))}
                  </div>
                  <div
                    className="tab-content"
                    role="tabpanel"
                    id="investigation-panel"
                    aria-labelledby={'tab-' + tab}
                  >
                    {investigation.error ? (
                      <ErrorMessage message={investigation.error.message} />
                    ) : null}
                    {tab === 'live' ? (
                      <Live
                        detail={investigation.data}
                        latest={detail.data.latest_run}
                        onStart={() => start.mutate('')}
                        starting={start.isPending}
                        onCancel={() => cancel.mutate()}
                        cancelling={cancel.isPending}
                        onReport={() => setTab('report')}
                      />
                    ) : tab === 'report' ? (
                      <ReportPanel
                        detail={investigation.data}
                        onOpen={setOpenEvidence}
                        onReview={() => setTab('review')}
                      />
                    ) : (
                      <ReviewPanel
                        detail={detail.data}
                        onChanged={invalidate}
                        onFollowup={(instruction) => start.mutate(instruction)}
                        busy={start.isPending}
                      />
                    )}
                  </div>
                </>
              ) : null}
            </div>
          </section>
          <footer className="main-footer">
            <span>
              <ShieldCheck size={13} /> Fakta dari data. Hipotesis dari AI. Keputusan dari manusia.
            </span>
            <span>{data?.employees ?? 150} pekerja sintetis · 60 hari riwayat</span>
          </footer>
        </main>
      </div>
      <EvidenceDialog evidence={openEvidence} onClose={() => setOpenEvidence(null)} />
    </div>
  )
}
function Stat({
  label,
  value,
  icon,
  note,
}: {
  label: string
  value?: React.ReactNode
  icon: React.ReactNode
  note: string
}) {
  return (
    <div className="stat-card">
      <div>
        <span>{label}</span>
        <i>{icon}</i>
      </div>
      <strong>{value ?? '—'}</strong>
      <small>{note}</small>
    </div>
  )
}
function Live({
  detail,
  latest,
  onStart,
  starting,
  onCancel,
  cancelling,
  onReport,
}: {
  detail?: RunDetail
  latest: Run | null
  onStart: () => void
  starting: boolean
  onCancel: () => void
  cancelling: boolean
  onReport: () => void
}) {
  const run = detail?.run ?? latest
  if (!run)
    return (
      <Empty title="Bukti siap ditelusuri">
        <p>Agent memilih tool, memeriksa konteks lintas departemen, lalu menyusun hipotesis.</p>
        <button className="primary" onClick={onStart} disabled={starting}>
          <Play size={16} />
          {starting ? 'Menambahkan ke antrean…' : 'Mulai investigasi'}
        </button>
      </Empty>
    )
  return (
    <>
      <div className="run-status">
        <span className={'run-icon ' + (active(run.status) ? 'pulse' : '')}>
          <Workflow size={23} />
        </span>
        <div>
          <h3>{active(run.status) ? 'Menelusuri bukti operasional' : stateLabel[run.status]}</h3>
          <p>
            {run.mode === 'mock' ? 'Mode simulasi · ' : 'Meta AI · '}
            {run.tool_calls} tool selesai · {run.model_rounds} putaran model
          </p>
        </div>
        <Badge value={run.status} />
      </div>
      {run.error ? <ErrorMessage message={run.error} /> : null}
      <div className="progress-track">
        {['finance', 'workforce', 'production', 'maintenance'].map((domain, i) => {
          const done = detail?.evidence.some((e) => e.domain === domain)
          return (
            <div className={done ? 'done' : ''} key={domain}>
              <span>{done ? <Check size={12} /> : i + 1}</span>
              {['Biaya', 'Workforce', 'Produksi', 'Maintenance'][i]}
            </div>
          )
        })}
      </div>
      <div className="section-heading">
        <h3>Jejak investigasi</h3>
        <span className="muted">
          {active(run.status) ? 'SSE · polling cadangan' : 'Tersimpan di PostgreSQL'}
        </span>
      </div>
      <ol className="timeline">
        {detail?.events.map((event) => (
          <li key={event.id}>
            <span className={'timeline-dot ' + event.kind} />
            <div>
              <strong>{event.message}</strong>
              <small>{timestamp(event.created_at)}</small>
            </div>
            {event.kind === 'evidence_collected' ? <ShieldCheck size={14} /> : null}
          </li>
        ))}
      </ol>
      <div className="run-actions">
        {active(run.status) ? (
          <button className="secondary" onClick={onCancel} disabled={cancelling}>
            <Square size={13} />
            Batalkan investigasi
          </button>
        ) : run.report ? (
          <button className="primary" onClick={onReport}>
            Buka laporan bukti <ArrowRight size={16} />
          </button>
        ) : (
          <button className="secondary" onClick={onStart} disabled={starting}>
            Investigasi ulang
          </button>
        )}
        <span className="muted">
          {'Pemakaian tercatat: US$' + (run.cost_micros / 1e6).toFixed(4)}
        </span>
      </div>
    </>
  )
}
function ReportPanel({
  detail,
  onOpen,
  onReview,
}: {
  detail?: RunDetail
  onOpen: (e: Evidence) => void
  onReview: () => void
}) {
  if (!detail?.run.report)
    return (
      <Empty title="Laporan belum tersedia">
        <p>Laporan muncul setelah agent selesai mengumpulkan dan memvalidasi bukti.</p>
      </Empty>
    )
  const report = detail.run.report
  return (
    <>
      <div className="report-intro">
        <span className="eyebrow">
          SINTESIS INVESTIGASI {detail.run.mode === 'mock' ? '· SIMULASI' : ''}
        </span>
        <p>{report.summary}</p>
        <small>
          Kesimpulan sementara memerlukan verifikasi manusia. Format dan referensi valid tidak
          membuktikan kausalitas.
        </small>
      </div>
      <div className="section-heading">
        <h3>Observasi dari catatan sumber</h3>
        <span className="verified-label">
          <ShieldCheck size={13} /> Dihitung oleh Go
        </span>
      </div>
      <div className="evidence-grid">
        {detail.evidence.map((e) => (
          <EvidenceCard key={e.id} evidence={e} onOpen={onOpen} />
        ))}
      </div>
      <div className="section-heading">
        <h3>Hipotesis kontribusi</h3>
        <span className="subtle-tag">BELUM TERBUKTI KAUSAL</span>
      </div>
      {report.hypotheses.map((h, i) => (
        <article className="hypothesis" key={i}>
          <span className="hypothesis-number">0{i + 1}</span>
          <div>
            <span className="confidence">
              Kekuatan bukti: {h.confidence === 'medium' ? 'sedang' : 'rendah'}
            </span>
            <p>{h.statement}</p>
            {h.counter_evidence.map((t) => (
              <small key={t}>{t}</small>
            ))}
            <div className="evidence-links">
              {h.evidence_ids.map((id) => {
                const e = detail.evidence.find((v) => v.id === id)
                return e ? (
                  <button key={id} onClick={() => onOpen(e)}>
                    <Layers3 size={12} />
                    {toolLabel[e.tool]}
                  </button>
                ) : null
              })}
            </div>
          </div>
        </article>
      ))}
      <div className="information-gaps">
        <h3>Yang masih perlu diketahui</h3>
        <ul>
          {report.information_gaps.map((g, i) => (
            <li key={i}>{g}</li>
          ))}
        </ul>
      </div>
      {detail.run.status === 'succeeded' ? (
        <button className="primary" onClick={onReview}>
          Tinjau rekomendasi <ArrowRight size={16} />
        </button>
      ) : null}
    </>
  )
}
function ReviewPanel({
  detail,
  onChanged,
  onFollowup,
  busy,
}: {
  detail: CaseDetail
  onChanged: () => void
  onFollowup: (instruction: string) => void
  busy: boolean
}) {
  const [reasons, setReasons] = useState<Record<string, string>>({}),
    [instruction, setInstruction] = useState('')
  const review = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: string }) =>
      api('/recommendations/' + id + '/reviews', { decision, reason: reasons[id] ?? '' }),
    onSuccess: onChanged,
  })
  const latest = detail.latest_run,
    recs = detail.recommendations.filter((r) => r.run_id === latest?.id)
  return (
    <>
      <div className="review-notice">
        <ShieldCheck size={19} />
        <div>
          <strong>Anda memegang keputusan akhir.</strong>
          <p>
            Persetujuan mencatat keputusan review. Data bisnis dan jadwal pekerja tidak berubah
            otomatis.
          </p>
        </div>
      </div>
      {review.error ? <ErrorMessage message={review.error.message} /> : null}
      {!recs.length ? (
        <Empty title="Belum ada rekomendasi untuk direview">
          <p>Investigasi yang berhasil menghasilkan tindakan beserta buktinya.</p>
        </Empty>
      ) : (
        recs.map((rec, i) => (
          <article className="recommendation" key={rec.id}>
            <div className="recommendation-top">
              <span className="action-number">{i + 1}</span>
              <div>
                <span className="eyebrow">{rec.department}</span>
                <h3>{rec.title}</h3>
              </div>
              {rec.decision ? (
                <Badge value={rec.decision} />
              ) : (
                <span className="subtle-tag">MENUNGGU ANDA</span>
              )}
            </div>
            <p>{rec.rationale}</p>
            <small className="muted">
              {rec.evidence_ids.length} bukti pendukung · run {rec.run_id.slice(0, 8)}
            </small>
            {rec.decision ? (
              <div className="review-record">
                <Check size={14} />
                {rec.reviewer} · {rec.reason || 'Disetujui tanpa catatan tambahan'}
              </div>
            ) : (
              <>
                <label htmlFor={'reason-' + rec.id}>
                  Catatan peninjau <span>(wajib untuk penolakan)</span>
                </label>
                <textarea
                  id={'reason-' + rec.id}
                  rows={2}
                  maxLength={2000}
                  value={reasons[rec.id] ?? ''}
                  onChange={(e) => setReasons((v) => ({ ...v, [rec.id]: e.target.value }))}
                  placeholder="Pertimbangan atau informasi tambahan…"
                />
                <div className="review-buttons">
                  <button
                    className="primary"
                    disabled={review.isPending || latest?.status !== 'succeeded'}
                    onClick={() => review.mutate({ id: rec.id, decision: 'approved' })}
                  >
                    <Check size={15} />
                    Setujui rekomendasi
                  </button>
                  <button
                    className="secondary"
                    disabled={
                      review.isPending ||
                      !(reasons[rec.id] ?? '').trim() ||
                      latest?.status !== 'succeeded'
                    }
                    onClick={() => review.mutate({ id: rec.id, decision: 'rejected' })}
                  >
                    Tolak
                  </button>
                </div>
              </>
            )}
          </article>
        ))
      )}
      <div className="followup">
        <h3>Masih ada pertanyaan?</h3>
        <label htmlFor="followup">Instruksi investigasi lanjutan</label>
        <textarea
          id="followup"
          maxLength={2000}
          rows={2}
          placeholder="Bukti apa yang perlu ditinjau kembali?"
          value={instruction}
          onChange={(e) => setInstruction(e.target.value)}
        />
        <button
          className="secondary"
          onClick={() => onFollowup(instruction)}
          disabled={busy || active(latest?.status) || !instruction.trim()}
        >
          <Search size={15} />
          Investigasi lanjutan
        </button>
      </div>
      <div className="section-heading">
        <h3>Jejak keputusan</h3>
        <span className="muted">Audit append-only</span>
      </div>
      <ol className="audit-list">
        {detail.audit.map((v) => (
          <li key={v.id}>
            <span className="status-dot" />
            <div>
              <strong>
                {(
                  {
                    case_detected: 'Anomali dideteksi',
                    investigation_queued: 'Investigasi dijadwalkan',
                    investigation_succeeded: 'Laporan selesai',
                    investigation_failed: 'Investigasi terhenti',
                    investigation_insufficient_evidence: 'Bukti belum cukup',
                    recommendation_reviewed: 'Keputusan review dicatat',
                    investigation_cancelled: 'Investigasi dibatalkan',
                  } as Record<string, string>
                )[v.action] ?? v.action}
                {v.details.decision ? ' · ' + stateLabel[v.details.decision] : ''}
              </strong>
              <small>
                {v.actor} · {timestamp(v.created_at)}
              </small>
            </div>
          </li>
        ))}
      </ol>
    </>
  )
}
