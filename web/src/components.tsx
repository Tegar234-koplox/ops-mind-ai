import { ApiError } from './api'
import { useI18n } from './i18n/context'
import { useEffect, useRef } from 'react'
import { ArrowUpRight, Check, FileText, LoaderCircle, X } from 'lucide-react'
import type { Evidence } from './types'
export function Badge({ value }: { value: string }) {
  const { t, stateLabel } = useI18n()

  return (
    <span className={'badge badge-' + value}>
      {stateLabel[value] ??
        { high: t('Prioritas tinggi'), medium: t('Prioritas sedang'), low: t('Prioritas rendah') }[
          value
        ] ??
        value}
    </span>
  )
}
export function Busy({ text }: { text?: string }) {
  const { t } = useI18n()

  return (
    <div className="busy" role="status">
      <LoaderCircle size={18} className="spin" />
      {text ?? t('Memuat data…')}
    </div>
  )
}
export function ErrorMessage({ message }: { message: string | Error }) {
  const { sourceText, locale, t } = useI18n()

  return (
    <div className="error-message" role="alert">
      {locale === 'en' &&
      message instanceof ApiError &&
      sourceText(message.message) === message.message
        ? t(
            message.code === 'validation'
              ? 'Permintaan tidak valid. Periksa isian dan coba lagi.'
              : 'Terjadi kesalahan layanan. Coba lagi setelah beberapa saat.',
          )
        : sourceText(typeof message === 'string' ? message : message.message)}
    </div>
  )
}
export function Empty({ title, children }: { title: string; children?: React.ReactNode }) {
  return (
    <div className="empty">
      <span className="empty-icon">
        <FileText size={26} />
      </span>
      <h3>{title}</h3>
      {children}
    </div>
  )
}
export function EvidenceCard({
  evidence,
  onOpen,
}: {
  evidence: Evidence
  onOpen: (e: Evidence) => void
}) {
  const { t, factLabel, factValue, toolLabel } = useI18n()

  return (
    <button className="evidence-card" onClick={() => onOpen(evidence)}>
      <div className="evidence-title">
        <span className="evidence-icon">
          <FileText size={16} />
        </span>
        <strong>{toolLabel[evidence.tool] ?? evidence.tool}</strong>
        <ArrowUpRight size={16} />
      </div>
      <dl>
        {Object.entries(evidence.facts)
          .filter(([k]) => !['currency', 'baseline_basis', 'aggregation'].includes(k))
          .map(([k, v]) => (
            <div key={k}>
              <dt>{factLabel[k] ?? k}</dt>
              <dd>{factValue(k, v)}</dd>
            </div>
          ))}
      </dl>
      <div className="evidence-foot">
        <span>
          <Check size={12} /> {evidence.sources.length} {t('referensi')}
        </span>
        <Badge value={evidence.completeness} />
      </div>
    </button>
  )
}
export function EvidenceDialog({
  evidence,
  onClose,
}: {
  evidence: Evidence | null
  onClose: () => void
}) {
  const { t, sourceText, factLabel, factValue, toolLabel, timestamp } = useI18n()

  const ref = useRef<HTMLDialogElement>(null)
  useEffect(() => {
    if (evidence) ref.current?.showModal()
    else ref.current?.close()
  }, [evidence])
  return (
    <dialog
      ref={ref}
      className="evidence-dialog"
      onCancel={onClose}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
      aria-labelledby="evidence-heading"
    >
      {evidence ? (
        <>
          <div className="dialog-head">
            <div>
              <span className="eyebrow">{t('Sumber terverifikasi')}</span>
              <h2 id="evidence-heading">{toolLabel[evidence.tool] ?? evidence.tool}</h2>
            </div>
            <button aria-label={t('Tutup bukti')} className="icon-button" onClick={onClose}>
              <X />
            </button>
          </div>
          <Badge value={evidence.completeness} />
          <dl className="evidence-facts">
            {Object.entries(evidence.facts).map(([k, v]) => (
              <div key={k}>
                <dt>{factLabel[k] ?? k}</dt>
                <dd>{factValue(k, v)}</dd>
              </div>
            ))}
          </dl>
          {evidence.notes.map((n) => (
            <p className="notice" key={n}>
              {sourceText(n)}
            </p>
          ))}
          <h3>{t('Jejak sumber')}</h3>
          <p className="muted">
            {t('Snapshot diambil')} {timestamp(evidence.retrieved_at)}
            {t('. Referensi catatan sintetis asli:')}
          </p>
          <div className="source-list">
            {evidence.sources.map((s) => (
              <div key={s.type + s.id}>
                <span>{s.type}</span>
                <code>{s.id}</code>
              </div>
            ))}
          </div>
          <div className="hash">
            <strong>{t('SHA-256 snapshot')}</strong>
            <code>{evidence.hash}</code>
          </div>
        </>
      ) : null}
    </dialog>
  )
}
