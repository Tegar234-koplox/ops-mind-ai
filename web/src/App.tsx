import type { Locale } from './i18n/format'
import { useI18n, LanguagePicker, ReportLanguage } from './i18n/context'
import { useEffect, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Activity,
  ArrowLeft,
  ArrowRight,
  Bell,
  Check,
  ChevronRight,
  CircleDot,
  ClipboardList,
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
import { active, api, ApiError } from './api'
import { Badge, Busy, Empty, ErrorMessage, EvidenceCard, EvidenceDialog } from './components'
import { HandoverWorkspace } from './handover/Workspace'
import type { Case, CaseDetail, Evidence, Overview, Run, RunDetail } from './types'
type Session = { username: string; model_mode: string }
type CaseTab = 'live' | 'report' | 'review'
export default function App() {
  const { t } = useI18n()

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
        <LanguagePicker />
        <Busy text={t('Menghubungkan ruang investigasi…')} />
      </div>
    )
  if (session.error)
    return (
      <div className="app-loading">
        <Brand />
        <LanguagePicker />
        <ErrorMessage message={session.error} />
        <button className="primary" onClick={() => session.refetch()}>
          {t('Coba lagi')}
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
  const { t } = useI18n()

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
          <span className="eyebrow light">{t('Intelijen operasi manufaktur')}</span>
          <h1>
            {t('Setiap anomali')}
            <br />
            {t('punya cerita.')}
            <br />
            <em>{t('Temukan buktinya.')}</em>
          </h1>
          <p>
            {t(
              'Investigasi lintas departemen yang menghubungkan data operasional dengan keputusan manusia.',
            )}
          </p>
          <div className="login-flow">
            <span>
              <Activity size={18} />
              {t('Deteksi')}
            </span>
            <ChevronRight size={14} />
            <span>
              <Search size={18} />
              {t('Investigasi')}
            </span>
            <ChevronRight size={14} />
            <span>
              <ShieldCheck size={18} />
              {t('Review')}
            </span>
          </div>
        </div>
        <div className="login-foot">
          <Factory size={16} />
          {t('Satu fasilitas. Bukti dari seluruh operasi.')}
        </div>
      </section>
      <section className="login-form-wrap">
        <div className="login-language">
          <LanguagePicker />
        </div>
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
          <span className="eyebrow">{t('Akses operator')}</span>
          <h2>{t('Masuk ke ruang investigasi')}</h2>
          <p className="muted">{t('Gunakan kredensial operator pada konfigurasi aplikasi.')}</p>
          <label htmlFor="username">{t('Nama pengguna')}</label>
          <input
            id="username"
            autoComplete="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <label htmlFor="password">{t('Kata sandi')}</label>
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          {login.error ? <ErrorMessage message={login.error} /> : null}
          <button className="primary" disabled={login.isPending}>
            {login.isPending ? t('Memeriksa akses…') : t('Masuk')}
            <ArrowRight size={17} />
          </button>
          <p className="login-disclaimer">
            <ShieldCheck size={15} />
            {t(
              'Data manufaktur sintetis. Keputusan operasional tetap memerlukan peninjau manusia.',
            )}
          </p>
        </form>
      </section>
    </main>
  )
}
function Workspace({ session }: { session: Session }) {
  const { t, sourceText, locale, factLabel, factValue, ruleLabel, stateLabel } = useI18n()

  const client = useQueryClient(),
    [view, setView] = useState<'investigations' | 'handover'>('investigations'),
    [selected, setSelected] = useState(''),
    [search, setSearch] = useState(''),
    [tabChoice, setTabChoice] = useState<{
      caseID: string
      runID?: string
      tab: CaseTab
    }>(),
    [mobileDetail, setMobileDetail] = useState(false),
    [openEvidence, setOpenEvidence] = useState<Evidence | null>(null),
    [error, setError] = useState<Error | null>(null)
  const casePanel = useRef<HTMLDivElement>(null)
  const inbox = useRef<HTMLDivElement>(null)
  const listScroll = useRef(0)
  const returningToList = useRef(false)
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
  const tab =
    tabChoice && tabChoice.caseID === caseID && tabChoice.runID === runID
      ? tabChoice.tab
      : detail.data?.latest_run?.report
        ? 'report'
        : 'live'
  function openTab(next: CaseTab) {
    if (caseID) setTabChoice({ caseID, runID, tab: next })
  }
  function openCase(id: string) {
    if (window.matchMedia('(max-width: 640px)').matches && !mobileDetail)
      listScroll.current = window.scrollY
    setSelected(id)
    setTabChoice(undefined)
    setMobileDetail(true)
    setView('investigations')
  }
  function showCases() {
    returningToList.current = mobileDetail && window.matchMedia('(max-width: 640px)').matches
    setMobileDetail(false)
    setView('investigations')
  }
  useEffect(() => {
    if (view !== 'investigations' || !window.matchMedia('(max-width: 640px)').matches) return
    if (mobileDetail) {
      casePanel.current?.focus({ preventScroll: true })
      casePanel.current?.scrollIntoView({ block: 'start' })
    } else if (returningToList.current) {
      returningToList.current = false
      window.scrollTo(0, listScroll.current)
      inbox.current
        ?.querySelector<HTMLButtonElement>('.case-item.active')
        ?.focus({ preventScroll: true })
    }
  }, [caseID, mobileDetail, view])
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
    mutationFn: (language: Locale) => api('/detections', { report_language: language }),
    onSuccess: () => {
      setError(null)
      invalidate()
    },
    onError: (e) => setError(e),
  })
  const start = useMutation({
    mutationFn: ({ instruction, language }: { instruction: string; language: Locale }) =>
      api<Run>(
        '/cases/' + caseID + (instruction ? '/follow-up-investigations' : '/investigations'),
        { instruction, report_language: language },
      ),
    onSuccess: async (run) => {
      const caseKey = ['case', run.case_id]
      await client.cancelQueries({ queryKey: caseKey })
      client.setQueryData<CaseDetail>(caseKey, (previous) =>
        previous ? { ...previous, latest_run: run, recommendations: [] } : previous,
      )
      client.setQueryData<RunDetail>(['run', run.id], { run, evidence: [], events: [] })
      if (caseID === run.case_id) setTabChoice(undefined)
      setError(null)
      invalidate()
    },
    onError: (e) => setError(e),
  })
  const cancel = useMutation({
    mutationFn: () => api('/investigations/' + runID + '/cancel', {}),
    onSuccess: invalidate,
    onError: (e) => setError(e),
  })
  const logout = useMutation({
    mutationFn: () => api('/auth/logout', {}),
    onSuccess: async () => {
      await client.cancelQueries()
      client.setQueryData(['session'], null)
      client.removeQueries({ predicate: (q) => q.queryKey[0] !== 'session' })
    },
    onError: (e) => setError(e),
  })
  const data = overview.data,
    visible = items.filter((c) =>
      (ruleLabel[c.rule_id] + ' ' + sourceText(c.department) + ' ' + c.rule_id)
        .toLowerCase()
        .includes(search.toLowerCase()),
    )
  const latest = detail.data?.latest_run
  const recommendations = detail.data?.recommendations.filter((rec) => rec.run_id === runID) ?? []
  const pendingReviews = recommendations.filter((rec) => !rec.decision).length
  return (
    <div
      className={
        'app-shell' +
        (view === 'handover' ? ' is-handover' : '') +
        (mobileDetail ? ' case-detail-open' : '')
      }
    >
      <aside className="sidebar">
        <Brand />
        <div className="workspace-label">{t('RUANG KERJA')}</div>
        <nav aria-label={t('Navigasi utama')}>
          <button
            className={'nav-item' + (view === 'investigations' ? ' selected' : '')}
            aria-current={view === 'investigations' ? 'page' : undefined}
            onClick={showCases}
          >
            <LayoutList size={18} /> {t('Kasus')} <span>{data?.cases ?? 0}</span>
          </button>
          <button
            className={'nav-item' + (view === 'handover' ? ' selected' : '')}
            aria-current={view === 'handover' ? 'page' : undefined}
            onClick={() => setView('handover')}
          >
            <ClipboardList size={18} />
            {t('Serah terima shift')}
          </button>
        </nav>
        <div className="sidebar-bottom">
          <div className="facility">
            <span>
              <Factory size={19} />
            </span>
            <div>
              <strong>{t('Fasilitas demo')}</strong>
              <small>FACTORY-01 · Asia/Bangkok</small>
            </div>
          </div>
          <div className="sidebar-note">
            <span className="status-dot" />
            {t('Data sintetis · Seed 20261027')}
          </div>
        </div>
      </aside>
      <div className="workspace">
        <header className="topbar">
          <div className="breadcrumb">
            {t('Operasional')}
            <ChevronRight size={13} />
            <strong>{view === 'handover' ? t('Serah terima shift') : t('Kasus')}</strong>
          </div>
          <div className="operator">
            <LanguagePicker />
            <span className="top-status">
              <span className="status-dot" />
              {t('Workspace terlindungi')}
            </span>
            <span className="avatar">{session.username[0].toUpperCase()}</span>
            <span>{session.username}</span>
            <button
              className="icon-button"
              aria-label={t('Keluar')}
              onClick={() => logout.mutate()}
            >
              <LogOut size={17} />
            </button>
          </div>
        </header>
        <main className="main-content">
          <div
            className="workspace-switch handover-screen-only"
            aria-label={t('Pilih ruang kerja')}
          >
            <button
              className="secondary"
              aria-pressed={view === 'investigations'}
              onClick={showCases}
            >
              {t('Kasus')}
            </button>
            <button
              className="secondary"
              aria-pressed={view === 'handover'}
              onClick={() => setView('handover')}
            >
              {t('Serah terima shift')}
            </button>
          </div>
          {view === 'handover' ? (
            <HandoverWorkspace onOpenCase={openCase} />
          ) : (
            <>
              <div className="page-heading case-overview">
                <div>
                  <span className="eyebrow">{t('INTELIJEN OPERASIONAL')}</span>
                  <h1>
                    {t('Ruang investigasi')}
                    <span>.</span>
                  </h1>
                  <p>{t('Hubungkan anomali, telusuri bukti, ambil keputusan.')}</p>
                </div>
                <button
                  className="primary"
                  onClick={() => detect.mutate(locale)}
                  disabled={detect.isPending}
                >
                  <Activity size={17} />
                  {detect.isPending ? t('Mendeteksi…') : t('Jalankan deteksi')}
                </button>
              </div>
              <div className={'mode-banner ' + (data?.model_mode === 'mock' ? 'mock' : 'meta')}>
                <Sparkles size={17} />
                <strong>{data?.model_mode === 'mock' ? t('SIMULASI MODEL') : t('META AI')}</strong>
                <span>
                  {data?.model_mode === 'mock'
                    ? t(
                        'Respons model tiruan untuk pengujian. Data, tool, dan penyimpanan tetap nyata.',
                      )
                    : (data?.model_name ?? t('Menghubungkan model')) +
                      t(' · Agent memilih bukti melalui tool baca terbatas.')}
                </span>
              </div>
              {error ? <ErrorMessage message={error} /> : null}
              {overview.error ? <ErrorMessage message={overview.error} /> : null}
              <section className="stat-grid case-overview" aria-label={t('Status workspace')}>
                <Stat
                  label={t('Kasus terdeteksi')}
                  value={data?.cases}
                  icon={<LayoutList />}
                  note={(data?.active_rule_count ?? 4) + t(' aturan aktif')}
                />
                <Stat
                  label={t('Investigasi aktif')}
                  value={data?.active_runs}
                  icon={<Activity />}
                  note={t('Proses dengan checkpoint')}
                />
                <Stat
                  label={t('Menunggu review')}
                  value={data?.pending_reviews}
                  icon={<Bell />}
                  note={t('Keputusan ada pada manusia')}
                />
                <Stat
                  label={t('Terpakai / direservasi')}
                  value={
                    data ? 'US$' + ((data.budget_micros.total ?? 0) / 1e6).toFixed(3) : undefined
                  }
                  icon={<CircleDot />}
                  note={t('Batas aplikasi US$') + (data?.credit_limit_micros ?? 20000000) / 1e6}
                />
              </section>
              <section className="investigation-workspace">
                <div className="inbox" ref={inbox}>
                  <div className="panel-heading">
                    <h2>
                      {t('Inbox kasus')}
                      <span>{items.length}</span>
                    </h2>
                    <span className="subtle-tag">{t('SEMUA LINI')}</span>
                  </div>
                  <label className="search">
                    <Search size={16} />
                    <input
                      aria-label={t('Cari kasus')}
                      placeholder={t('Cari kasus atau departemen…')}
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                    />
                  </label>
                  <div className="case-list">
                    {cases.isPending ? (
                      <Busy />
                    ) : cases.error ? (
                      <ErrorMessage message={cases.error} />
                    ) : visible.length ? (
                      visible.map((c) => (
                        <button
                          key={c.id}
                          className={'case-item ' + (c.id === caseID ? 'active' : '')}
                          aria-pressed={c.id === caseID}
                          onClick={() => openCase(c.id)}
                        >
                          <div className="case-item-top">
                            <span className={'case-rule ' + c.severity}>{c.rule_id}</span>
                            <span className="case-department">{sourceText(c.department)}</span>
                          </div>
                          <h3>{ruleLabel[c.rule_id] ?? c.rule_id}</h3>
                          <p>
                            {t('Lini perakitan')} {c.entity_id} {t('· Shift operasional')}
                          </p>
                          <div className="case-item-bottom">
                            <Badge value={c.severity} />
                            <span>{stateLabel[c.status] ?? c.status}</span>
                          </div>
                        </button>
                      ))
                    ) : (
                      <Empty title={search ? t('Kasus tidak ditemukan') : t('Belum ada kasus')}>
                        <p>
                          {search
                            ? t('Gunakan kata kunci lain.')
                            : t('Jalankan deteksi untuk memeriksa lini produksi.')}
                        </p>
                      </Empty>
                    )}
                  </div>
                  <div className="inbox-foot">
                    <ShieldCheck size={13} />
                    {t('Seluruh catatan adalah data sintetis')}
                  </div>
                </div>
                <div
                  className="case-panel"
                  ref={casePanel}
                  tabIndex={-1}
                  role="region"
                  aria-label={t('Detail kasus')}
                >
                  <button className="secondary back-to-cases" onClick={showCases}>
                    <ArrowLeft size={16} /> {t('Kembali ke daftar kasus')}
                  </button>
                  {!caseID ? (
                    <Empty title={t('Mulai dari satu pertanyaan')}>
                      <p>
                        {t('Mengapa biaya lembur meningkat saat output menurun?')}
                        <br />
                        {t('Jalankan deteksi untuk menghubungkan buktinya.')}
                      </p>
                      <button
                        className="secondary"
                        disabled={detect.isPending}
                        onClick={() => detect.mutate(locale)}
                      >
                        {t('Deteksi anomali')}
                        <ArrowRight size={15} />
                      </button>
                    </Empty>
                  ) : detail.isPending ? (
                    <Busy text={t('Memuat konteks kasus…')} />
                  ) : detail.error ? (
                    <ErrorMessage message={detail.error} />
                  ) : detail.data ? (
                    <>
                      <div className="case-header">
                        <div>
                          <span className="eyebrow">
                            {t('LINI PERAKITAN')} {detail.data.case.entity_id} <span> / </span>{' '}
                            {detail.data.case.rule_id}
                          </span>
                          <h2>{ruleLabel[detail.data.case.rule_id] ?? detail.data.case.rule_id}</h2>
                          <p>{sourceText(detail.data.case.explanation)}</p>
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
                      {latest?.report ? (
                        <div className="case-next-step">
                          <div>
                            <strong>
                              {latest.status === 'insufficient_evidence'
                                ? stateLabel[latest.status]
                                : pendingReviews
                                  ? t('Investigasi selesai — menunggu review')
                                  : recommendations.length
                                    ? t('Review selesai')
                                    : t('Investigasi selesai')}
                            </strong>
                            <p>
                              {latest.status === 'insufficient_evidence'
                                ? t('Bukti belum cukup untuk mengambil keputusan.')
                                : pendingReviews
                                  ? pendingReviews + ' ' + t('rekomendasi menunggu keputusan Anda.')
                                  : t('Laporan dan bukti tersimpan siap ditelusuri.')}
                            </p>
                          </div>
                          {tab !== 'report' ? (
                            <button className="secondary" onClick={() => openTab('report')}>
                              {t('Buka laporan bukti')} <ArrowRight size={16} />
                            </button>
                          ) : recommendations.length ? (
                            <button className="primary" onClick={() => openTab('review')}>
                              {pendingReviews ? t('Tinjau rekomendasi') : t('Lihat hasil review')}{' '}
                              <ArrowRight size={16} />
                            </button>
                          ) : latest.status === 'insufficient_evidence' ? (
                            <button className="secondary" onClick={() => openTab('live')}>
                              {t('Tinjau proses investigasi')}
                            </button>
                          ) : null}
                        </div>
                      ) : null}
                      <div className="tabs" role="tablist" aria-label={t('Tahap investigasi')}>
                        {(
                          [
                            ['live', t('Investigasi langsung')],
                            ['report', t('Laporan bukti')],
                            ['review', t('Review manusia')],
                          ] as const
                        ).map(([id, label], i) => (
                          <button
                            role="tab"
                            id={'tab-' + id}
                            aria-controls="investigation-panel"
                            aria-selected={tab === id}
                            tabIndex={tab === id ? 0 : -1}
                            key={id}
                            className={tab === id ? 'selected' : ''}
                            onClick={() => openTab(id)}
                            onKeyDown={(event) => {
                              const tabs: CaseTab[] = ['live', 'report', 'review']
                              const next =
                                event.key === 'ArrowRight'
                                  ? (i + 1) % 3
                                  : event.key === 'ArrowLeft'
                                    ? (i + 2) % 3
                                    : event.key === 'Home'
                                      ? 0
                                      : event.key === 'End'
                                        ? 2
                                        : -1
                              if (next < 0) return
                              event.preventDefault()
                              openTab(tabs[next])
                              document.getElementById('tab-' + tabs[next])?.focus()
                            }}
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
                          <ErrorMessage message={investigation.error} />
                        ) : null}
                        {runID && investigation.isPending ? (
                          <Busy />
                        ) : tab === 'live' ? (
                          <Live
                            detail={investigation.data}
                            latest={detail.data.latest_run}
                            onStart={() => start.mutate({ instruction: '', language: locale })}
                            starting={start.isPending}
                            onCancel={() => cancel.mutate()}
                            cancelling={cancel.isPending}
                          />
                        ) : tab === 'report' ? (
                          <ReportPanel detail={investigation.data} onOpen={setOpenEvidence} />
                        ) : (
                          <ReviewPanel
                            detail={detail.data}
                            onChanged={invalidate}
                            onFollowup={(instruction) =>
                              start.mutate({ instruction, language: locale })
                            }
                            busy={start.isPending}
                            evidence={investigation.data?.evidence ?? []}
                            onOpen={setOpenEvidence}
                          />
                        )}
                      </div>
                    </>
                  ) : null}
                </div>
              </section>
            </>
          )}
          <footer className="main-footer">
            <span>
              <ShieldCheck size={13} />
              {t('Fakta dari data. Hipotesis dari AI. Keputusan dari manusia.')}
            </span>
            <span>
              {data?.employees ?? 150} {t('pekerja sintetis · 60 hari riwayat')}
            </span>
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
}: {
  detail?: RunDetail
  latest: Run | null
  onStart: () => void
  starting: boolean
  onCancel: () => void
  cancelling: boolean
}) {
  const { t, sourceText, stateLabel, timestamp } = useI18n()

  const run = detail?.run ?? latest
  if (!run)
    return (
      <Empty title={t('Bukti siap ditelusuri')}>
        <p>
          {t('Agent memilih tool, memeriksa konteks lintas departemen, lalu menyusun hipotesis.')}
        </p>
        <button className="primary" onClick={onStart} disabled={starting}>
          <Play size={16} />
          {starting ? t('Menambahkan ke antrean…') : t('Mulai investigasi')}
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
          <h3>{active(run.status) ? t('Menelusuri bukti operasional') : stateLabel[run.status]}</h3>
          <p>
            {run.mode === 'mock' ? t('Mode simulasi · ') : t('Meta AI · ')}
            {run.tool_calls} {t('tool selesai ·')} {run.model_rounds} {t('putaran model')}
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
              {[t('Biaya'), t('Workforce'), t('Produksi'), t('Maintenance')][i]}
            </div>
          )
        })}
      </div>
      <details className="process-details" open={active(run.status)}>
        <summary>{t('Jejak investigasi')}</summary>
        <span className="muted">
          {active(run.status) ? t('SSE · polling cadangan') : t('Tersimpan di PostgreSQL')}
        </span>
        <ol className="timeline">
          {detail?.events.map((event) => (
            <li key={event.id}>
              <span className={'timeline-dot ' + event.kind} />
              <div>
                <strong>{sourceText(event.message)}</strong>
                <small>{timestamp(event.created_at)}</small>
              </div>
              {event.kind === 'evidence_collected' ? <ShieldCheck size={14} /> : null}
            </li>
          ))}
        </ol>
      </details>
      <div className="run-actions">
        {active(run.status) ? (
          <button className="secondary" onClick={onCancel} disabled={cancelling}>
            <Square size={13} />
            {t('Batalkan investigasi')}
          </button>
        ) : !run.report ? (
          <button className="secondary" onClick={onStart} disabled={starting}>
            {t('Investigasi ulang')}
          </button>
        ) : null}
        <span className="muted">
          {t('Pemakaian tercatat: US$') + (run.cost_micros / 1e6).toFixed(4)}
        </span>
      </div>
    </>
  )
}
function ReportPanel({ detail, onOpen }: { detail?: RunDetail; onOpen: (e: Evidence) => void }) {
  const { t, toolLabel } = useI18n()

  if (!detail?.run.report)
    return (
      <Empty title={t('Laporan belum tersedia')}>
        <p>{t('Laporan muncul setelah agent selesai mengumpulkan dan memvalidasi bukti.')}</p>
      </Empty>
    )
  const report = detail.run.report
  return (
    <>
      <ReportLanguage language={detail.run.report_language} />
      <div className="report-intro">
        <span className="eyebrow">
          {t('SINTESIS INVESTIGASI')} {detail.run.mode === 'mock' ? t('· SIMULASI') : ''}
        </span>
        <p lang={detail.run.report_language || 'id'}>{report.summary}</p>
        <small>
          {t(
            'Kesimpulan sementara memerlukan verifikasi manusia. Format dan referensi valid tidak membuktikan kausalitas.',
          )}
        </small>
      </div>
      <div className="section-heading">
        <h3>{t('Observasi dari catatan sumber')}</h3>
        <span className="verified-label">
          <ShieldCheck size={13} />
          {t('Dihitung oleh Go')}
        </span>
      </div>
      <div className="evidence-grid">
        {detail.evidence.map((e) => (
          <EvidenceCard key={e.id} evidence={e} onOpen={onOpen} />
        ))}
      </div>
      <div className="section-heading">
        <h3>{t('Hipotesis kontribusi')}</h3>
        <span className="subtle-tag">{t('BELUM TERBUKTI KAUSAL')}</span>
      </div>
      {report.hypotheses.map((h, i) => (
        <article className="hypothesis" key={i}>
          <span className="hypothesis-number">0{i + 1}</span>
          <div>
            <span className="confidence">
              {t('Kekuatan bukti:')} {h.confidence === 'medium' ? t('sedang') : t('rendah')}
            </span>
            <p lang={detail.run.report_language || 'id'}>{h.statement}</p>
            {h.counter_evidence.map((t) => (
              <small lang={detail.run.report_language || 'id'} key={t}>
                {t}
              </small>
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
        <h3>{t('Yang masih perlu diketahui')}</h3>
        <ul>
          {report.information_gaps.map((g, i) => (
            <li lang={detail.run.report_language || 'id'} key={i}>
              {g}
            </li>
          ))}
        </ul>
      </div>
    </>
  )
}
function ReviewPanel({
  detail,
  onChanged,
  onFollowup,
  busy,
  evidence,
  onOpen,
}: {
  detail: CaseDetail
  onChanged: () => void
  onFollowup: (instruction: string) => void
  busy: boolean
  evidence: Evidence[]
  onOpen: (e: Evidence) => void
}) {
  const { t, sourceText, stateLabel, timestamp, toolLabel } = useI18n()

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
          <strong>{t('Anda memegang keputusan akhir.')}</strong>
          <p>
            {t(
              'Persetujuan mencatat keputusan review. Data bisnis dan jadwal pekerja tidak berubah otomatis.',
            )}
          </p>
        </div>
      </div>
      {review.error ? <ErrorMessage message={review.error} /> : null}
      {!recs.length ? (
        <Empty title={t('Belum ada rekomendasi untuk direview')}>
          <p>{t('Investigasi yang berhasil menghasilkan tindakan beserta buktinya.')}</p>
        </Empty>
      ) : (
        recs.map((rec, i) => (
          <article className="recommendation" key={rec.id}>
            <div className="recommendation-top">
              <span className="action-number">{i + 1}</span>
              <div>
                <span className="eyebrow">{sourceText(rec.department)}</span>
                <h3 lang={latest?.report_language || 'id'}>{rec.title}</h3>
              </div>
              {rec.decision ? (
                <Badge value={rec.decision} />
              ) : (
                <span className="subtle-tag">{t('MENUNGGU ANDA')}</span>
              )}
            </div>
            <p lang={latest?.report_language || 'id'}>{rec.rationale}</p>
            <small className="muted">
              {rec.evidence_ids.length} {t('bukti pendukung · run')} {rec.run_id.slice(0, 8)}
            </small>
            <div className="evidence-links" aria-label={t('Bukti pendukung rekomendasi')}>
              {rec.evidence_ids.map((id) => {
                const item = evidence.find((e) => e.id === id && e.run_id === rec.run_id)
                return item ? (
                  <button key={id} onClick={() => onOpen(item)}>
                    <Layers3 size={14} />
                    {toolLabel[item.tool] ?? item.tool}
                  </button>
                ) : null
              })}
            </div>
            {rec.decision ? (
              <div className="review-record">
                <Check size={14} />
                {rec.reviewer} · {rec.reason || t('Disetujui tanpa catatan tambahan')}
              </div>
            ) : (
              <>
                <label htmlFor={'reason-' + rec.id}>
                  {t('Catatan peninjau')}
                  <span>{t('(wajib untuk penolakan)')}</span>
                </label>
                <textarea
                  id={'reason-' + rec.id}
                  rows={2}
                  maxLength={2000}
                  value={reasons[rec.id] ?? ''}
                  onChange={(e) => setReasons((v) => ({ ...v, [rec.id]: e.target.value }))}
                  placeholder={t('Pertimbangan atau informasi tambahan…')}
                />
                <div className="review-buttons">
                  <button
                    className="primary"
                    disabled={review.isPending || latest?.status !== 'succeeded'}
                    onClick={() => review.mutate({ id: rec.id, decision: 'approved' })}
                  >
                    <Check size={15} />
                    {t('Setujui rekomendasi')}
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
                    {t('Tolak')}
                  </button>
                </div>
              </>
            )}
          </article>
        ))
      )}
      <p className="muted">
        {t('Laporan AI baru menggunakan bahasa aplikasi saat investigasi dimulai.')}
      </p>
      <div className="followup">
        <h3>{t('Masih ada pertanyaan?')}</h3>
        <label htmlFor="followup">{t('Instruksi investigasi lanjutan')}</label>
        <textarea
          id="followup"
          maxLength={2000}
          rows={2}
          placeholder={t('Bukti apa yang perlu ditinjau kembali?')}
          value={instruction}
          onChange={(e) => setInstruction(e.target.value)}
        />
        <button
          className="secondary"
          onClick={() => onFollowup(instruction)}
          disabled={busy || active(latest?.status) || !instruction.trim()}
        >
          <Search size={15} />
          {t('Investigasi lanjutan')}
        </button>
      </div>
      <div className="section-heading">
        <h3>{t('Jejak keputusan')}</h3>
        <span className="muted">{t('Audit append-only')}</span>
      </div>
      <ol className="audit-list">
        {detail.audit.map((v) => (
          <li key={v.id}>
            <span className="status-dot" />
            <div>
              <strong>
                {(
                  {
                    case_detected: t('Anomali dideteksi'),
                    investigation_queued: t('Investigasi dijadwalkan'),
                    investigation_succeeded: t('Laporan selesai'),
                    investigation_failed: t('Investigasi terhenti'),
                    investigation_insufficient_evidence: t('Bukti belum cukup'),
                    recommendation_reviewed: t('Keputusan review dicatat'),
                    investigation_cancelled: t('Investigasi dibatalkan'),
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
