package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"strconv"
	"time"
)

var ToolNames = []string{"get_shift_staffing", "get_attendance_summary", "get_approved_leave", "get_overtime_variance", "get_cost_breakdown", "get_production_performance", "get_downtime_events", "get_evidence_record"}

type ToolArgs struct {
	ShiftID    string `json:"shift_id"`
	EvidenceID string `json:"evidence_id,omitempty"`
}

func ParseArgs(name, raw, scope string) (ToolArgs, error) {
	var a ToolArgs
	known := false
	for _, n := range ToolNames {
		if n == name {
			known = true
		}
	}
	if !known {
		return a, errors.New("tool tidak diizinkan")
	}
	d := json.NewDecoder(bytes.NewBufferString(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&a); e != nil {
		return a, errors.New("parameter tool tidak valid")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return a, errors.New("parameter JSON ganda")
	}
	if a.ShiftID != scope {
		return a, errors.New("shift di luar scope kasus")
	}
	if name == "get_evidence_record" && a.EvidenceID == "" {
		return a, errors.New("evidence_id wajib")
	}
	if name != "get_evidence_record" && a.EvidenceID != "" {
		return a, errors.New("evidence_id hanya untuk get_evidence_record")
	}
	return a, nil
}
func Definitions() []map[string]any {
	out := []map[string]any{}
	for _, n := range ToolNames {
		props := map[string]any{"shift_id": map[string]any{"type": "string", "description": "Harus sama dengan shift_id kasus."}}
		required := []string{"shift_id"}
		if n == "get_evidence_record" {
			props["evidence_id"] = map[string]any{"type": "string"}
			required = append(required, "evidence_id")
		}
		out = append(out, map[string]any{"type": "function", "function": map[string]any{"name": n, "description": map[string]string{"get_shift_staffing": "Rekonsiliasi roster, hadir terverifikasi, cuti dan status belum jelas.", "get_attendance_summary": "Kelengkapan punch; pending tidak membuktikan mangkir.", "get_approved_leave": "Cuti disetujui untuk mengecualikan absensi yang sah.", "get_overtime_variance": "Biaya lembur aktual dan median empat pembanding, dihitung Go.", "get_cost_breakdown": "Pengeluaran lengkap per kategori pada shift.", "get_production_performance": "Target dan output aktual beserta shortfall dari Go.", "get_downtime_events": "Downtime union tanpa penghitungan ganda overlap.", "get_evidence_record": "Buka ulang snapshot bukti yang sudah dikumpulkan."}[n], "strict": true, "parameters": map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}}})
	}
	return out
}
func EvidenceHash(v domain.Evidence) string {
	b := store.JSON(map[string]any{"shift_id": v.ShiftID, "tool": v.Tool, "domain": v.Domain, "facts": v.Facts, "sources": v.Sources, "completeness": v.Completeness, "notes": v.Notes})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func ExecuteTool(ctx context.Context, s *store.Store, c domain.Case, run domain.Run, name string, a ToolArgs) (domain.Evidence, error) {
	v := domain.Evidence{ID: domain.ID(), CaseID: c.ID, RunID: run.ID, ShiftID: c.ShiftID, Tool: name, Facts: map[string]string{}, Sources: []domain.SourceRef{}, Completeness: "complete", Notes: []string{}, RetrievedAt: time.Now().UTC()}
	sh, e := s.Shift(ctx, c.ShiftID)
	if e != nil {
		return v, e
	}
	n := func(v int64) string { return strconv.FormatInt(v, 10) }
	switch name {
	case "get_shift_staffing", "get_attendance_summary":
		staff, e := s.Staffing(ctx, sh.ID)
		if e != nil {
			return v, e
		}
		v.Domain = "workforce"
		v.Sources = staff.Sources
		v.Facts = map[string]string{"scheduled": n(staff.Scheduled), "verified_present": n(staff.Present), "approved_leave": n(staff.Leave), "unverified": n(staff.Unverified), "required_staff": n(sh.Required)}
		if staff.Unverified > 0 {
			v.Completeness = "needs_verification"
			v.Notes = append(v.Notes, "Catatan belum terverifikasi tidak membuktikan mangkir; minta verifikasi supervisor.")
		}
	case "get_approved_leave":
		v.Domain = "workforce"
		rows, e := s.Pool.Query(ctx, "SELECT DISTINCT l.id FROM leave_records l JOIN shift_assignments a ON a.employee_id=l.employee_id WHERE a.shift_id=$1 AND l.approved AND l.starts_at<=$2 AND l.ends_at>=$3 ORDER BY l.id", sh.ID, sh.Start, sh.End)
		if e != nil {
			return v, e
		}
		for rows.Next() {
			var id string
			if e = rows.Scan(&id); e != nil {
				rows.Close()
				return v, e
			}
			v.Sources = append(v.Sources, domain.SourceRef{Type: "leave_records", ID: id})
		}
		err := rows.Err()
		rows.Close()
		if err != nil {
			return v, err
		}
		v.Facts["approved_leave"] = fmt.Sprint(len(v.Sources))
		v.Notes = append(v.Notes, "Cuti disetujui adalah ketidakhadiran yang sah, bukan pelanggaran.")
	case "get_overtime_variance":
		v.Domain = "finance"
		ot, e := s.Overtime(ctx, sh)
		if e != nil {
			return v, e
		}
		v.Sources = ot.Sources
		v.Facts = map[string]string{"currency": "IDR", "observed_rupiah": n(ot.Observed), "baseline_rupiah": n(ot.Baseline), "variance_bps": n(ot.Variance), "comparable_shifts": fmt.Sprint(ot.Comparable), "baseline_basis": "median empat shift, dinormalisasi target terencana"}
		if !ot.Complete {
			v.Completeness = "insufficient_evidence"
			v.Notes = append(v.Notes, "Baseline atau catatan lembur tidak cukup lengkap.")
		}
	case "get_cost_breakdown":
		v.Domain = "finance"
		rows, e := s.Pool.Query(ctx, "SELECT id,category,amount_rupiah,complete FROM expenses WHERE shift_id=$1 ORDER BY id", sh.ID)
		if e != nil {
			return v, e
		}
		totals := map[string]int64{}
		for rows.Next() {
			var id, category string
			var amount int64
			var complete bool
			if e = rows.Scan(&id, &category, &amount, &complete); e != nil {
				rows.Close()
				return v, e
			}
			totals[category] += amount
			v.Sources = append(v.Sources, domain.SourceRef{Type: "expenses", ID: id})
			if !complete {
				v.Completeness = "needs_verification"
			}
		}
		err := rows.Err()
		rows.Close()
		if err != nil {
			return v, err
		}
		v.Facts["currency"] = "IDR"
		for k, x := range totals {
			v.Facts[k+"_rupiah"] = n(x)
		}
	case "get_production_performance":
		v.Domain = "production"
		var pid string
		if e = s.Pool.QueryRow(ctx, "SELECT id FROM production_runs WHERE shift_id=$1", sh.ID).Scan(&pid); e != nil {
			return v, e
		}
		v.Sources = []domain.SourceRef{{Type: "production_runs", ID: pid}}
		v.Facts = map[string]string{"planned_output": n(sh.Planned), "actual_output": n(sh.Actual)}
		if sh.Planned > 0 {
			bps, e := domain.BPS(sh.Actual, sh.Planned)
			if e != nil {
				return v, e
			}
			v.Facts["shortfall_bps"] = n(-bps)
		} else {
			v.Completeness = "insufficient_evidence"
		}
		if !sh.Complete {
			v.Completeness = "needs_verification"
		}
	case "get_downtime_events":
		v.Domain = "maintenance"
		minutes, src, e := s.Downtime(ctx, sh)
		if e != nil {
			return v, e
		}
		v.Sources = src
		v.Facts = map[string]string{"line_downtime_minutes": n(minutes), "aggregation": "union interval pada lini; overlap hanya dihitung sekali"}
	case "get_evidence_record":
		old, e := s.Evidence(ctx, c.ID, a.EvidenceID)
		if e != nil {
			return v, e
		}
		if old.ShiftID != c.ShiftID || old.RunID != run.ID {
			return v, errors.New("evidence di luar run/scope")
		}
		return old, nil
	default:
		return v, errors.New("tool tidak diizinkan")
	}
	if e = s.ValidateSources(ctx, c.ShiftID, v.Sources); e != nil {
		return v, e
	}
	v.Hash = EvidenceHash(v)
	return v, nil
}
func (s *Runner) cachedTool(ctx context.Context, runID, callID, argsHash string) (*domain.Evidence, error) {
	var raw []byte
	var hash string
	e := s.Store.Pool.QueryRow(ctx, "SELECT result,args_hash FROM tool_executions WHERE run_id=$1 AND call_id=$2", runID, callID).Scan(&raw, &hash)
	if e != nil {
		if normalizeNotFound(e) {
			return nil, nil
		}
		return nil, e
	}
	if hash != argsHash {
		return nil, errors.New("call_id digunakan dengan argumen berbeda")
	}
	var v domain.Evidence
	if e = json.Unmarshal(raw, &v); e != nil {
		return nil, e
	}
	return &v, nil
}
func (s *Runner) saveTool(ctx context.Context, run domain.Run, call Call, argsHash string, v domain.Evidence, ms int64) error {
	tx, e := s.Store.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "INSERT INTO case_evidence(id,case_id,run_id,shift_id,tool,domain,facts,sources,completeness,notes,hash,retrieved_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT(id) DO NOTHING", v.ID, v.CaseID, v.RunID, v.ShiftID, v.Tool, v.Domain, store.JSON(v.Facts), store.JSON(v.Sources), v.Completeness, store.JSON(v.Notes), v.Hash, v.RetrievedAt); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO tool_executions(id,run_id,call_id,name,args,args_hash,status,evidence_id,result,duration_ms) VALUES($1,$2,$3,$4,$5,$6,'succeeded',$7,$8,$9)", domain.ID(), run.ID, call.ID, call.Function.Name, []byte(call.Function.Arguments), argsHash, v.ID, store.JSON(v), ms); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO investigation_events(run_id,kind,message) VALUES($1,'evidence_collected',$2)", run.ID, "Bukti tersimpan: "+call.Function.Name); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
