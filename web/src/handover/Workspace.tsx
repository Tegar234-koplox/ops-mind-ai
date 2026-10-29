import { useState } from 'react'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { FileCheck2, Plus, Printer } from 'lucide-react'
import { api, timestamp } from '../api'
import { Busy, Empty, ErrorMessage } from '../components'
import { HandoverReport } from './Report'
import type { Catalogue, Handover, HandoverSummary } from './types'
import './styles.css'

const pageSize = 30

export function HandoverWorkspace({ onOpenCase }: { onOpenCase: (id: string) => void }) {
  const client = useQueryClient()
  const [line, setLine] = useState('')
  const [date, setDate] = useState('')
  const [shiftOffset, setShiftOffset] = useState(0)
  const [selectedShift, setSelectedShift] = useState('')
  const [historyOffset, setHistoryOffset] = useState(0)
  const [selectedReport, setSelectedReport] = useState({ shift: '', id: '' })
  const catalogue = useQuery({
    queryKey: ['handover-shifts', line, date, shiftOffset],
    placeholderData: keepPreviousData,
    queryFn: () =>
      api<Catalogue>(
        '/shifts?' +
          new URLSearchParams({
            line_id: line,
            date,
            limit: String(pageSize),
            offset: String(shiftOffset),
          }),
      ),
  })
  const shifts = catalogue.isPlaceholderData ? [] : (catalogue.data?.items ?? [])
  const shiftID = shifts.find((shift) => shift.id === selectedShift)?.id ?? shifts[0]?.id ?? ''
  const history = useQuery({
    queryKey: ['handovers', shiftID, historyOffset],
    queryFn: () =>
      api<{ items: HandoverSummary[]; total: number }>(
        '/handovers?' +
          new URLSearchParams({
            shift_id: shiftID,
            limit: String(pageSize),
            offset: String(historyOffset),
          }),
      ),
    enabled: !!shiftID,
  })
  const reportID =
    selectedReport.shift === shiftID &&
    history.data?.items.some((item) => item.id === selectedReport.id)
      ? selectedReport.id
      : (history.data?.items[0]?.id ?? '')
  const report = useQuery({
    queryKey: ['handover', reportID],
    queryFn: () => api<Handover>('/handovers/' + reportID),
    enabled: !!reportID,
  })
  const create = useMutation({
    mutationFn: (id: string) => api<Handover>('/handovers', { shift_id: id }),
    onSuccess: async (result) => {
      await client.cancelQueries({ queryKey: ['handovers', result.shift_id] })
      client.setQueryData(['handover', result.id], result)
      client.setQueryData<{ items: HandoverSummary[]; total: number }>(
        ['handovers', result.shift_id, 0],
        (previous) =>
          previous
            ? {
                items: [result, ...previous.items.filter((item) => item.id !== result.id)].slice(
                  0,
                  pageSize,
                ),
                total:
                  previous.total + (previous.items.some((item) => item.id === result.id) ? 0 : 1),
              }
            : previous,
      )
      if (result.shift_id === shiftID) {
        setSelectedReport({ shift: result.shift_id, id: result.id })
        setHistoryOffset(0)
      }
      void client.invalidateQueries({ queryKey: ['handovers', result.shift_id] })
    },
  })
  function resetSelection() {
    setSelectedShift('')
    setShiftOffset(0)
    setHistoryOffset(0)
    setSelectedReport({ shift: '', id: '' })
  }

  return (
    <section className="handover-workspace">
      <div className="page-heading handover-screen-only">
        <div>
          <span className="eyebrow">LAPORAN OPERASIONAL</span>
          <h1>
            Serah terima shift<span>.</span>
          </h1>
          <p>Teruskan fakta, catatan yang belum jelas, dan keputusan manusia.</p>
        </div>
        <FileCheck2 size={30} className="handover-heading-icon" aria-hidden="true" />
      </div>
      <div className="handover-controls handover-screen-only">
        <div className="handover-filters">
          <div>
            <label htmlFor="handover-line">Lini produksi</label>
            <select
              id="handover-line"
              value={line}
              onChange={(event) => {
                setLine(event.target.value)
                resetSelection()
              }}
            >
              <option value="">Semua lini</option>
              {catalogue.data?.lines.map((item) => (
                <option value={item.id} key={item.id}>
                  {item.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="handover-date">Tanggal operasional</label>
            <input
              id="handover-date"
              type="date"
              value={date}
              onChange={(event) => {
                setDate(event.target.value)
                resetSelection()
              }}
            />
          </div>
          <div>
            <label htmlFor="handover-shift">Shift selesai</label>
            <select
              id="handover-shift"
              value={shiftID}
              disabled={!shifts.length || catalogue.isFetching}
              onChange={(event) => {
                setSelectedShift(event.target.value)
                setHistoryOffset(0)
                setSelectedReport({ shift: '', id: '' })
              }}
            >
              {!shifts.length ? (
                <option value="">Tidak ada shift</option>
              ) : (
                shifts.map((shift) => (
                  <option key={shift.id} value={shift.id}>
                    {shift.line_name} · {shift.operational_date} · {shift.code}
                  </option>
                ))
              )}
            </select>
          </div>
        </div>
        <div className="handover-toolbar">
          <div>
            <button
              className="primary"
              disabled={!shiftID || create.isPending || catalogue.isFetching}
              onClick={() => create.mutate(shiftID)}
            >
              <Plus size={16} />
              {create.isPending ? 'Menyusun…' : 'Susun laporan baru'}
            </button>
            {report.data ? (
              <button className="secondary" onClick={() => window.print()}>
                <Printer size={16} />
                Cetak / Simpan PDF
              </button>
            ) : null}
          </div>
          <div className="handover-pagination">
            <button
              className="secondary"
              disabled={shiftOffset === 0 || catalogue.isFetching}
              onClick={() => {
                setShiftOffset((value) => Math.max(0, value - pageSize))
                setHistoryOffset(0)
              }}
            >
              Shift lebih baru
            </button>
            <button
              className="secondary"
              disabled={
                shiftOffset + pageSize >= (catalogue.data?.total ?? 0) || catalogue.isFetching
              }
              onClick={() => {
                setShiftOffset((value) => value + pageSize)
                setHistoryOffset(0)
              }}
            >
              Shift lebih lama
            </button>
          </div>
        </div>
        <p className="muted">
          {catalogue.data ? 'Evaluasi data: ' + timestamp(catalogue.data.as_of) + '. ' : ''}
          Penyusunan memakai fakta Go dan investigasi tersimpan, tanpa panggilan model tambahan.
          Setiap versi disimpan terpisah.
        </p>
      </div>
      {catalogue.error ? <ErrorMessage message={catalogue.error.message} /> : null}
      {create.error && create.variables === shiftID ? (
        <ErrorMessage
          message={'Laporan ' + (create.variables ?? '') + ': ' + create.error.message}
        />
      ) : null}
      {history.error ? <ErrorMessage message={history.error.message} /> : null}
      {report.error ? <ErrorMessage message={report.error.message} /> : null}
      {catalogue.isPending || catalogue.isPlaceholderData ? (
        <Busy text="Memuat shift…" />
      ) : !shiftID ? (
        <Empty title="Tidak ada shift selesai pada pilihan ini">
          <p>Ubah lini atau tanggal operasional.</p>
        </Empty>
      ) : (
        <>
          <div className="handover-history handover-screen-only">
            <div>
              <label htmlFor="handover-history">Versi laporan tersimpan</label>
              <select
                id="handover-history"
                value={reportID}
                disabled={!history.data?.items.length || history.isFetching}
                onChange={(event) => setSelectedReport({ shift: shiftID, id: event.target.value })}
              >
                {!history.data?.items.length ? (
                  <option value="">Belum ada laporan</option>
                ) : (
                  history.data.items.map((item) => (
                    <option key={item.id} value={item.id}>
                      {timestamp(item.created_at)} · {item.created_by} · {item.id.slice(0, 8)}
                    </option>
                  ))
                )}
              </select>
            </div>
            <div className="handover-pagination">
              <button
                className="secondary"
                disabled={historyOffset === 0 || history.isFetching}
                onClick={() => setHistoryOffset((value) => Math.max(0, value - pageSize))}
              >
                Versi lebih baru
              </button>
              <button
                className="secondary"
                disabled={
                  historyOffset + pageSize >= (history.data?.total ?? 0) || history.isFetching
                }
                onClick={() => setHistoryOffset((value) => value + pageSize)}
              >
                Versi lebih lama
              </button>
            </div>
          </div>
          {history.isPending || (reportID && report.isPending) ? (
            <Busy text="Memuat laporan tersimpan…" />
          ) : report.data ? (
            <HandoverReport report={report.data} onOpenCase={onOpenCase} />
          ) : (
            <Empty title="Belum ada laporan serah terima">
              <p>Susun laporan untuk menyimpan fakta dan keputusan pada shift ini.</p>
            </Empty>
          )}
        </>
      )}
    </section>
  )
}
