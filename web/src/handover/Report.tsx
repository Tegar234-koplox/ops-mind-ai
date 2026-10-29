import { ArrowUpRight, FileCheck2, Fingerprint } from 'lucide-react'
import { Badge } from '../components'
import { factLabel, factValue, ruleLabel, stateLabel, timestamp } from '../api'
import type { Handover } from './types'

export function HandoverReport({
  report,
  onOpenCase,
}: {
  report: Handover
  onOpenCase: (id: string) => void
}) {
  const snapshot = report.snapshot
  return (
    <article className="handover-report" aria-label="Laporan serah terima tersimpan">
      <header className="handover-report-header">
        <div>
          <span className="eyebrow">LAPORAN SERAH TERIMA · DATA SINTETIS</span>
          <h2>
            {snapshot.shift.line_name} · {snapshot.shift.code}
          </h2>
          <p>
            Tanggal operasional {snapshot.shift.operational_date} · {snapshot.shift.product}
          </p>
          <p>
            {timestamp(snapshot.shift.starts_at)} — {timestamp(snapshot.shift.ends_at)}
          </p>
        </div>
        <FileCheck2 size={30} aria-hidden="true" />
      </header>
      <dl className="handover-metadata">
        <div>
          <dt>Disusun oleh</dt>
          <dd>
            {report.created_by} · {timestamp(report.created_at)}
          </dd>
        </div>
        <div>
          <dt>Evaluasi data operasional</dt>
          <dd>{timestamp(snapshot.data_as_of)}</dd>
        </div>
        <div>
          <dt>Shift berikutnya</dt>
          <dd>
            {snapshot.next_shift
              ? snapshot.next_shift.code + ' · ' + timestamp(snapshot.next_shift.starts_at)
              : 'Belum tersedia'}
          </dd>
        </div>
        <div>
          <dt>ID laporan</dt>
          <dd>{report.id}</dd>
        </div>
      </dl>
      <div className="handover-notice">
        <strong>Snapshot untuk ditinjau manusia</strong>
        <p>
          Laporan mencatat kondisi dan keputusan saat disusun. Tidak ada penugasan, persetujuan,
          atau tindakan bisnis yang dijalankan otomatis.
        </p>
      </div>
      <section className="handover-facts" aria-label="Fakta operasional dari Go">
        {snapshot.sections.map((section) => (
          <div className="handover-fact-section" key={section.key}>
            <div className="handover-section-title">
              <h3>{section.title}</h3>
              <Badge value={section.quality} />
            </div>
            <dl>
              {Object.entries(section.facts).map(([key, value]) => (
                <div key={key}>
                  <dt>{factLabel[key] ?? key}</dt>
                  <dd>{factValue(key, value)}</dd>
                </div>
              ))}
            </dl>
            <ul>
              {section.notes.map((note) => (
                <li key={note}>{note}</li>
              ))}
            </ul>
          </div>
        ))}
      </section>
      <section className="handover-section" aria-labelledby="handover-machines">
        <h3 id="handover-machines">Catatan gangguan mesin</h3>
        {snapshot.downtime_events.length ? (
          <div className="handover-machine-list">
            {snapshot.downtime_events.map((event) => (
              <div key={event.id}>
                <strong>{event.machine_name}</strong>
                <p>
                  {timestamp(event.starts_at)} — {timestamp(event.ends_at)}
                </p>
                <p>{event.reason}</p>
                <small>{event.id}</small>
              </div>
            ))}
          </div>
        ) : (
          <p>
            Tidak ada interval downtime tercatat. Kondisi mesin tetap perlu diperiksa saat serah
            terima.
          </p>
        )}
      </section>
      <section className="handover-section" aria-labelledby="handover-cases">
        <h3 id="handover-cases">Kasus, rekomendasi, dan keputusan tersimpan</h3>
        {!snapshot.cases.length ? (
          <p>
            Belum ada kasus tersimpan. Jalankan deteksi pada ruang investigasi bila diperlukan;
            ketiadaan kasus bukan bukti operasi bebas anomali.
          </p>
        ) : (
          snapshot.cases.map((item) => (
            <article className="handover-case" key={item.case_id}>
              <div className="handover-section-title">
                <h4>
                  {item.rule_id} · {ruleLabel[item.rule_id] ?? item.rule_id}
                </h4>
                <Badge value={item.status} />
              </div>
              <p>{item.explanation}</p>
              {item.synthesis ? (
                <div className="handover-synthesis">
                  <strong>
                    {item.model_mode === 'mock'
                      ? 'SIMULASI · Sintesis investigasi tersimpan'
                      : 'Sintesis AI dari investigasi tersimpan'}
                  </strong>
                  <p>{item.synthesis}</p>
                </div>
              ) : (
                <p className="muted">
                  {item.run_status
                    ? 'Status investigasi: ' + (stateLabel[item.run_status] ?? item.run_status)
                    : 'Belum ada investigasi tersimpan.'}
                </p>
              )}
              <ul className="handover-recommendations">
                {item.recommendations.map((rec) => (
                  <li key={rec.id}>
                    <div className="handover-section-title">
                      <strong>{rec.title}</strong>
                      <Badge value={rec.decision || 'review_required'} />
                    </div>
                    <p>{rec.rationale}</p>
                    {rec.decision ? (
                      <small>
                        {rec.reviewer} · {rec.reason || 'Disetujui tanpa catatan tambahan'}
                      </small>
                    ) : (
                      <small>Menunggu review manusia; belum mengotorisasi tindakan.</small>
                    )}
                  </li>
                ))}
              </ul>
              <button
                className="secondary handover-screen-only"
                onClick={() => onOpenCase(item.case_id)}
              >
                Telusuri kasus {item.rule_id}
                <ArrowUpRight size={14} />
              </button>
              <small className="handover-reference">
                {item.evidence.length} bukti tersimpan{item.run_id ? ' · run ' + item.run_id : ''}
              </small>
            </article>
          ))
        )}
      </section>
      <section className="handover-section handover-limitations">
        <h3>Hal yang perlu diperhatikan penerima</h3>
        <ul>
          {snapshot.notes.map((note) => (
            <li key={note}>{note}</li>
          ))}
        </ul>
      </section>
      <footer className="handover-provenance">
        <strong>
          <Fingerprint size={15} /> SHA-256 snapshot
        </strong>
        <code>{report.hash}</code>
        <p>
          Hash memverifikasi konsistensi snapshot, bukan membuktikan seluruh kesimpulan benar.{' '}
          {snapshot.sources.length} referensi sumber tersimpan.
        </p>
        <details className="handover-screen-only">
          <summary>Lihat seluruh referensi sumber</summary>
          <ul>
            {snapshot.sources.map((source) => (
              <li key={source.type + ':' + source.id}>
                <code>
                  {source.type} · {source.id}
                </code>
              </li>
            ))}
          </ul>
        </details>
      </footer>
    </article>
  )
}
