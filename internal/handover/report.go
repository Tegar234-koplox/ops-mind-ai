package handover

import (
	"context"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"strconv"
	"time"
)

type Service struct{ Store *store.Store }

func (s Service) Create(ctx context.Context, shiftID, actor string) (domain.Handover, error) {
	data, err := s.Store.HandoverData(ctx, shiftID, domain.AsOf)
	if err != nil {
		return domain.Handover{}, err
	}
	report := domain.Handover{
		ID:        domain.ID(),
		ShiftID:   shiftID,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		CreatedBy: actor,
		Snapshot:  Build(data),
	}
	if report.Hash, err = domain.HandoverHash(report.Snapshot); err != nil {
		return report, err
	}
	err = s.Store.SaveHandover(ctx, report)
	return report, err
}

// Build formats facts computed by Go. It never invokes a model or changes business records.
func Build(data store.HandoverData) domain.HandoverSnapshot {
	snapshot := domain.HandoverSnapshot{
		Version:        1,
		FacilityID:     domain.FacilityID,
		Timezone:       "Asia/Bangkok",
		Synthetic:      true,
		DataAsOf:       data.AsOf,
		Shift:          data.Shift,
		NextShift:      data.NextShift,
		Sections:       []domain.HandoverSection{workforce(data), production(data), finance(data), maintenance(data)},
		Cases:          data.Cases,
		DowntimeEvents: data.DowntimeEvents,
		Sources:        data.Sources,
		Notes: []string{
			"Laporan adalah snapshot saat disusun. Gunakan versi baru untuk melihat koreksi sumber atau keputusan review berikutnya.",
			"Data manufaktur sintetis; waktu evaluasi data mengikuti fixture, bukan kondisi pabrik saat ini.",
			"Kasus hanya berasal dari deteksi tersimpan. Laporan tidak menjalankan deteksi, investigasi baru, atau tindakan bisnis.",
			"Sintesis AI yang disertakan berasal dari investigasi tersimpan; korelasi bukan pembuktian sebab.",
		},
	}
	if len(data.Cases) == 0 {
		snapshot.Notes = append(snapshot.Notes, "Belum ada kasus tersimpan pada shift ini; ini bukan bukti bahwa operasi bebas anomali.")
	}
	if data.NextShift == nil {
		snapshot.Notes = append(snapshot.Notes, "Data shift berikutnya yang berurutan belum tersedia; jangan mengasumsikan penugasan atau penerima laporan.")
	}
	return snapshot
}

func section(key, title string) domain.HandoverSection {
	return domain.HandoverSection{Key: key, Title: title, Quality: "complete", Facts: map[string]string{}, Notes: []string{}}
}
func number(value int64) string { return strconv.FormatInt(value, 10) }
func verify(s *domain.HandoverSection, note string) {
	s.Quality = "needs_verification"
	s.Notes = append(s.Notes, note)
}

func workforce(data store.HandoverData) domain.HandoverSection {
	s := section("workforce", "Kehadiran dan cakupan operator")
	v := data.Staff
	s.Facts = map[string]string{
		"scheduled":        number(v.Scheduled),
		"verified_present": number(v.Present),
		"approved_leave":   number(v.Leave),
		"confirmed_absent": number(v.ConfirmedAbsent),
		"unverified":       number(v.Unverified),
		"required_staff":   number(data.Operation.Required),
	}
	if v.Unverified > 0 {
		verify(&s, "Verifikasi catatan kehadiran yang belum jelas bersama supervisor. Catatan pending bukan bukti mangkir.")
	}
	if v.Leave > 0 {
		s.Notes = append(s.Notes, "Cuti disetujui merupakan ketidakhadiran yang sah dan tidak dinilai sebagai pelanggaran.")
	}
	if v.Present < data.Operation.Required {
		s.Notes = append(s.Notes, "Tinjau cakupan operator pada shift berikutnya; selisih terhadap kebutuhan tidak membuktikan penyebab kekurangan output.")
	}
	if !data.Shift.Working {
		s.Notes = append(s.Notes, "Tanggal operasional termasuk hari nonkerja; periksa kalender dan penugasan aktual sebelum menilai kehadiran.")
	}
	return s
}
func production(data store.HandoverData) domain.HandoverSection {
	s := section("production", "Target dan hasil produksi")
	v := data.Operation
	s.Facts = map[string]string{"planned_output": number(v.Planned), "actual_output": number(v.Actual)}
	if !v.Complete {
		verify(&s, "Catatan shift atau produksi belum lengkap; angka masih memerlukan verifikasi.")
	}
	bps, err := domain.BPS(v.Actual, v.Planned)
	if err != nil {
		verify(&s, "Target nol atau tidak valid; persentase kekurangan output tidak dapat dihitung.")
	} else {
		shortfall := int64(0)
		if bps < 0 {
			shortfall = -bps
		}
		s.Facts["shortfall_bps"] = number(shortfall)
	}
	if v.Actual < v.Planned {
		s.Notes = append(s.Notes, "Tinjau kekurangan output dan catatan operasional bersama lini penerima; hubungan dengan staffing atau downtime belum membuktikan sebab.")
	}
	return s
}
func finance(data store.HandoverData) domain.HandoverSection {
	s := section("finance", "Lembur dan pengeluaran")
	ot, cost, approval := data.Overtime, data.Expenses, data.Approval
	s.Facts = map[string]string{
		"observed_rupiah":             number(ot.Observed),
		"expense_actual_rupiah":       number(cost.Observed),
		"expense_posting_closed":      strconv.FormatBool(cost.PostingClosed),
		"unapproved_records":          number(approval.Unapproved),
		"unapproved_minutes":          number(approval.Minutes),
		"unverified_overtime_records": number(approval.Unverified),
	}
	if ot.Complete {
		s.Facts["baseline_rupiah"] = number(ot.Baseline)
		s.Facts["variance_bps"] = number(ot.Variance)
	} else {
		verify(&s, "Baseline atau sumber lembur belum cukup lengkap; jangan menyimpulkan persentase kenaikan.")
	}
	if cost.Complete {
		s.Facts["expense_baseline_rupiah"] = number(cost.Baseline)
		s.Facts["expense_variance_bps"] = number(cost.Variance)
	} else {
		verify(&s, "Posting biaya belum ditutup atau baseline/sumber biaya belum lengkap. Total tercatat belum membuktikan seluruh pengeluaran final.")
	}
	if approval.Unverified > 0 {
		verify(&s, "Sebagian catatan lembur belum terverifikasi; status persetujuan perlu diperiksa.")
	}
	if approval.Unapproved > 0 {
		s.Notes = append(s.Notes, "Tinjau lembur tanpa persetujuan efektif bersama penanggung jawab; riwayat pembatalan persetujuan tetap menjadi bukti.")
	}
	s.Notes = append(s.Notes, "Baseline dinormalisasi target terencana, bukan output aktual. Biaya rendah bukan otomatis efisiensi; anomali biaya bukan bukti fraud.")
	return s
}
func maintenance(data store.HandoverData) domain.HandoverSection {
	s := section("maintenance", "Gangguan mesin dan downtime lini")
	s.Facts = map[string]string{"downtime_minutes": number(data.Downtime), "downtime_event_count": strconv.Itoa(len(data.DowntimeEvents))}
	s.Notes = append(s.Notes, "Durasi lini dihitung sebagai union interval yang dibatasi shift; interval mesin yang bertumpang tindih tidak dijumlahkan ganda.")
	if data.Downtime > 0 {
		s.Notes = append(s.Notes, "Periksa kondisi mesin dan catatan gangguan sebelum serah terima. Berakhirnya interval downtime tidak membuktikan perbaikan telah selesai.")
	}
	return s
}
