package detection

import (
	"context"
	"fmt"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"time"
)

type shiftRule func(context.Context, store.Shift, time.Time) (*domain.Case, error)

func (e Engine) p1Cases(ctx context.Context, shift store.Shift, asOf time.Time) ([]domain.Case, error) {
	cases := []domain.Case{}
	rules := []shiftRule{e.absenceStreakCase, e.unapprovedOvertimeCase, e.expenseVarianceCase}
	for _, evaluate := range rules {
		result, err := evaluate(ctx, shift, asOf)
		if err != nil {
			return nil, err
		}
		if result != nil {
			cases = append(cases, *result)
		}
	}
	return cases, nil
}

func p1Case(shift store.Shift, rule, department string) domain.Case {
	return domain.Case{
		ID: domain.ID(), RuleID: rule, EntityID: shift.LineID, ShiftID: shift.ID,
		Department: department, Severity: "medium", Status: "open", DataQuality: "confirmed",
	}
}

func (e Engine) absenceStreakCase(ctx context.Context, shift store.Shift, asOf time.Time) (*domain.Case, error) {
	streaks, err := e.Store.AbsenceStreaks(ctx, shift, asOf)
	if err != nil {
		return nil, err
	}
	if len(streaks.Employees) == 0 {
		return nil, nil
	}
	result := p1Case(shift, "ATT-001", "Workforce")
	result.Expected = map[string]string{"required_consecutive_shifts": "3"}
	result.Observed = map[string]string{
		"confirmed_absence_workers": fmt.Sprint(len(streaks.Employees)),
		"consecutive_shifts":        "3",
		"employee_ids":              store.EmployeeIDs(streaks),
	}
	result.Explanation = "Tiga shift kerja terjadwal terakhir tercatat absen terkonfirmasi; cuti memutus rangkaian. Ini bukan kesimpulan pelanggaran pekerja."
	result.Sources = streaks.Sources
	return &result, nil
}

func (e Engine) unapprovedOvertimeCase(ctx context.Context, shift store.Shift, asOf time.Time) (*domain.Case, error) {
	approvals, err := e.Store.OvertimeApprovals(ctx, shift, asOf)
	if err != nil {
		return nil, err
	}
	if approvals.Unapproved == 0 && approvals.PendingApproval == 0 {
		return nil, nil
	}
	result := p1Case(shift, "OT-002", "Keuangan")
	if approvals.Unverified > 0 {
		result.DataQuality = "needs_verification"
	}
	result.Expected = map[string]string{"unapproved_records": "0"}
	result.Observed = map[string]string{
		"unapproved_records":            store.Num(approvals.Unapproved),
		"unapproved_minutes":            store.Num(approvals.Minutes),
		"unverified_overtime_records":   store.Num(approvals.Unverified),
		"unverified_unapproved_records": store.Num(approvals.PendingApproval),
	}
	result.Explanation = "Lembur aktual tercatat tanpa persetujuan yang berlaku pada waktu evaluasi; riwayat pembatalan persetujuan tetap disimpan."
	result.Sources = approvals.Sources
	return &result, nil
}

func (e Engine) expenseVarianceCase(ctx context.Context, shift store.Shift, asOf time.Time) (*domain.Case, error) {
	costs, err := e.Store.ExpenseVariance(ctx, shift, asOf)
	if err != nil {
		return nil, err
	}
	if !costs.Complete || (costs.Variance <= e.Config.CostThreshold && -costs.Variance <= e.Config.CostThreshold) {
		return nil, nil
	}
	rule := "COST-001"
	message := "Pengeluaran melebihi baseline empat shift sebanding. Kenaikan biaya bukan bukti pemborosan atau fraud."
	if costs.Variance < 0 {
		rule = "COST-002"
		message = "Pengeluaran di bawah baseline dengan posting telah ditutup. Selisih ini belum membuktikan efisiensi."
	}
	result := p1Case(shift, rule, "Keuangan")
	result.Expected = map[string]string{
		"expense_baseline_rupiah": store.Num(costs.Baseline),
		"threshold_bps":           store.Num(e.Config.CostThreshold),
	}
	result.Observed = map[string]string{
		"expense_actual_rupiah": store.Num(costs.Observed),
		"expense_variance_bps":  store.Num(costs.Variance),
		"comparable_shifts":     fmt.Sprint(costs.Comparable),
	}
	result.Explanation = message
	result.Sources = costs.Sources
	if costs.Variance >= 3500 || costs.Variance <= -3500 {
		result.Severity = "high"
	}
	return &result, nil
}
