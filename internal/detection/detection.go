package detection

import (
	"context"
	"fmt"
	"opsmind/internal/config"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"time"
)

type Engine struct {
	Store  *store.Store
	Config config.Config
}

func (e Engine) Execute(ctx context.Context, shiftID string, asOf time.Time) ([]domain.Case, error) {
	ids := []string{}
	if shiftID != "" {
		ids = append(ids, shiftID)
	} else {
		rows, err := e.Store.Pool.Query(ctx, "SELECT id FROM shift_instances WHERE ends_at<=$1 ORDER BY ends_at DESC,id LIMIT 9", asOf)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			ids = append(ids, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
	}
	out := []domain.Case{}
	for _, id := range ids {
		sh, err := e.Store.Shift(ctx, id)
		if err != nil {
			return nil, err
		}
		if sh.End.After(asOf) || !sh.Working {
			continue
		}
		makeCase := func(rule, dept, quality string, expected, observed map[string]string, explanation string, src []domain.SourceRef) domain.Case {
			return domain.Case{ID: domain.ID(), RuleID: rule, EntityID: sh.LineID, ShiftID: id, Department: dept, Severity: "medium", Status: "open", DataQuality: quality, Expected: expected, Observed: observed, Explanation: explanation, Sources: src}
		}
		cases := []domain.Case{}
		ot, err := e.Store.Overtime(ctx, sh)
		if err != nil {
			return nil, err
		}
		if !ot.Complete {
			c := makeCase("OT-001", "Keuangan", "insufficient_evidence", map[string]string{"minimum_comparable_shifts": "4"}, map[string]string{"comparable_shifts": fmt.Sprint(ot.Comparable)}, "Baseline lembur belum cukup lengkap untuk mengonfirmasi anomali.", ot.Sources)
			c.Severity = "low"
			cases = append(cases, c)
		} else if ot.Variance > e.Config.OTThreshold {
			c := makeCase("OT-001", "Keuangan", "confirmed", map[string]string{"overtime_rupiah": store.Num(ot.Baseline), "threshold_bps": store.Num(e.Config.OTThreshold)}, map[string]string{"overtime_rupiah": store.Num(ot.Observed), "variance_bps": store.Num(ot.Variance)}, "Biaya lembur melebihi baseline empat shift sebanding; penyebab memerlukan investigasi.", ot.Sources)
			if ot.Variance >= 3500 {
				c.Severity = "high"
			}
			cases = append(cases, c)
		}
		if sh.Planned > 0 {
			bps, err := domain.BPS(sh.Actual, sh.Planned)
			if err != nil {
				return nil, err
			}
			if -bps > e.Config.ProdThreshold || !sh.Complete {
				quality := "confirmed"
				if !sh.Complete {
					quality = "needs_verification"
				}
				c := makeCase("PROD-001", "Produksi", quality, map[string]string{"output": store.Num(sh.Planned), "threshold_bps": store.Num(e.Config.ProdThreshold)}, map[string]string{"output": store.Num(sh.Actual), "shortfall_bps": store.Num(-bps)}, "Output berada di bawah target terencana pada shift yang selesai.", []domain.SourceRef{{Type: "production_runs", ID: "PR-" + id}})
				if -bps >= 2000 {
					c.Severity = "high"
				}
				cases = append(cases, c)
			}
		}
		staffing, err := e.Store.Staffing(ctx, id)
		if err != nil {
			return nil, err
		}
		if staffing.Unverified > 0 && !sh.End.Add(time.Duration(e.Config.AttendanceGrace)*time.Minute).After(asOf) {
			c := makeCase("ATT-002", "Workforce", "needs_verification", map[string]string{"complete_required_punches": store.Num(staffing.Scheduled - staffing.Leave - staffing.ConfirmedAbsent)}, map[string]string{"unverified": store.Num(staffing.Unverified), "approved_leave": store.Num(staffing.Leave), "verified_present": store.Num(staffing.Present)}, "Punch belum lengkap atau terverifikasi; tidak menyimpulkan mangkir atau pelanggaran.", staffing.Sources)
			c.Severity = "low"
			cases = append(cases, c)
		}
		downtime, src, err := e.Store.Downtime(ctx, sh)
		if err != nil {
			return nil, err
		}
		if downtime > e.Config.DowntimeThreshold {
			c := makeCase("MAINT-001", "Maintenance", "confirmed", map[string]string{"threshold_minutes": store.Num(e.Config.DowntimeThreshold)}, map[string]string{"downtime_minutes": store.Num(downtime)}, "Downtime gabungan lini melebihi threshold; interval overlap dihitung satu kali.", src)
			if downtime >= 120 {
				c.Severity = "high"
			}
			cases = append(cases, c)
		}
		if e.Config.EnableP1 {
			advancedCases, err := e.p1Cases(ctx, sh, asOf)
			if err != nil {
				return nil, err
			}
			cases = append(cases, advancedCases...)
		}
		for _, c := range cases {
			saved, created, err := e.Store.InsertCase(ctx, c, e.Config.Mode)
			if err != nil {
				return nil, err
			}
			if created || shiftID != "" {
				out = append(out, saved)
			}
		}
	}
	return out, nil
}
