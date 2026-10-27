package seed

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"math/rand"
	"opsmind/internal/store"
	"time"
)

const GoldenShift = "A-20261026-S1"

func Run(ctx context.Context, s *store.Store) (bool, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return false, e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(2026102701)"); e != nil {
		return false, e
	}
	var count int
	if e = tx.QueryRow(ctx, "SELECT count(*) FROM employees").Scan(&count); e != nil {
		return false, e
	}
	if count > 0 {
		return false, nil
	}
	batch := &pgx.Batch{}
	for _, d := range []struct{ id, name string }{{"OPS", "Operasi"}, {"HR", "Workforce"}, {"FIN", "Keuangan"}, {"MNT", "Maintenance"}} {
		batch.Queue("INSERT INTO departments VALUES($1,$2)", d.id, d.name)
	}
	for i := 1; i <= 150; i++ {
		dept, role := "OPS", "operator"
		if i > 120 {
			role = "support"
			dept = []string{"HR", "FIN", "MNT"}[(i-121)%3]
		}
		batch.Queue("INSERT INTO employees VALUES($1,$2,$3)", fmt.Sprintf("EMP-%03d", i), dept, role)
	}
	machine := 1
	for _, line := range []string{"A", "B", "C"} {
		batch.Queue("INSERT INTO production_lines(id,department_id,name) VALUES($1,'OPS',$2)", line, "Assembly Line "+line)
		n := 3
		if line == "C" {
			n = 2
		}
		for j := 0; j < n; j++ {
			batch.Queue("INSERT INTO machines VALUES($1,$2,$3)", fmt.Sprintf("M-%02d", machine), line, fmt.Sprintf("Mesin %02d", machine))
			machine++
		}
	}
	loc := time.FixedZone("Asia/Bangkok", 7*3600)
	start := time.Date(2026, 8, 28, 0, 0, 0, 0, loc)
	rng := rand.New(rand.NewSource(20261027))
	for day := 0; day < 60; day++ {
		date := start.AddDate(0, 0, day)
		working := date.Weekday() != time.Sunday
		batch.Queue("INSERT INTO calendar_days VALUES($1,$2,$3)", date.Format("2006-01-02"), working, map[bool]string{true: "Hari kerja sintetis", false: "Libur fasilitas sintetis"}[working])
		if !working {
			continue
		}
		operatorOffset := 0
		for shift := 1; shift <= 3; shift++ {
			for li, line := range []string{"A", "B", "C"} {
				id := fmt.Sprintf("%s-%s-S%d", line, date.Format("20060102"), shift)
				n := 8
				if line == "A" {
					n = 10
					if shift == 1 {
						n = 38
					}
				}
				begin := date.Add(time.Duration(6+(shift-1)*8) * time.Hour)
				end := begin.Add(8 * time.Hour)
				target := 400
				if line == "A" {
					target = 600
					if shift == 1 {
						target = 1000
					}
				}
				actual := target - rng.Intn(21)
				if id == GoldenShift {
					actual = 830
				}
				batch.Queue("INSERT INTO shift_instances VALUES($1,$2,$3,$4,$5,$6,$7,'WIDGET-X',true)", id, line, date.Format("2006-01-02"), fmt.Sprintf("S%d", shift), begin, end, n)
				batch.Queue("INSERT INTO production_runs VALUES($1,$2,$3,$4,true)", "PR-"+id, id, target, actual)
				otTotal := int64(300000)
				if line == "A" && shift == 1 {
					otTotal = 1000000
				}
				if id == GoldenShift {
					otTotal = 1380000
				}
				eligible := n
				if id == GoldenShift {
					eligible = 31
				}
				for j := 0; j < n; j++ {
					emp := fmt.Sprintf("EMP-%03d", 1+(operatorOffset+j+day*7)%120)
					assignment := fmt.Sprintf("AS-%s-%02d", id, j)
					batch.Queue("INSERT INTO shift_assignments VALUES($1,$2,$3)", assignment, id, emp)
					var cin, cout any = begin, end
					verified, complete := true, true
					if id == GoldenShift && j >= 31 {
						cin = nil
						cout = nil
						verified = false
						if j >= 34 {
							complete = false
						} else {
							batch.Queue("INSERT INTO leave_records VALUES($1,$2,$3,$4,true)", fmt.Sprintf("LEAVE-%s-%d", id, j), emp, begin, end)
						}
					}
					batch.Queue("INSERT INTO attendance_records VALUES($1,$2,$3,$4,$5,$6)", "AT-"+assignment, assignment, cin, cout, verified, complete)
					if j < eligible {
						minutes := int(otTotal/500) / eligible
						if j == 0 {
							minutes += int(otTotal/500) % eligible
						}
						cost := int64(minutes) * 500
						batch.Queue("INSERT INTO overtime_records VALUES($1,$2,$3,$4,true,true)", "OT-"+assignment, assignment, minutes, cost)
					}
				}
				operatorOffset += n
				batch.Queue("INSERT INTO expenses VALUES($1,$2,$3,'overtime',$4,$5,true)", "EX-OT-"+id, line, id, otTotal, end)
				batch.Queue("INSERT INTO expenses VALUES($1,$2,$3,'maintenance',$4,$5,true)", "EX-MT-"+id, line, id, int64(100000+rng.Intn(10000)), end)
				if id == GoldenShift {
					// [08:00,09:10] and [09:00,09:50] => union 110 minutes.
					batch.Queue("INSERT INTO downtime_events VALUES($1,$2,'M-01',$3,$4,$5)", "DT-GOLD-1", id, begin.Add(2*time.Hour), begin.Add(190*time.Minute), "Interupsi conveyor; menunggu verifikasi maintenance")
					batch.Queue("INSERT INTO downtime_events VALUES($1,$2,'M-02',$3,$4,$5)", "DT-GOLD-2", id, begin.Add(3*time.Hour), begin.Add(230*time.Minute), "Kalibrasi station; interval beririsan")
				} else if rng.Intn(4) == 0 {
					machineID := []string{"M-01", "M-04", "M-07"}[li]
					batch.Queue("INSERT INTO downtime_events VALUES($1,$2,$3,$4,$5,$6)", "DT-"+id, id, machineID, begin.Add(time.Hour), begin.Add(time.Duration(65+rng.Intn(20))*time.Minute), "Pemeriksaan singkat sintetis")
				}
			}
		}
	}
	for _, line := range []string{"A", "B", "C"} {
		batch.Queue("INSERT INTO budgets VALUES($1,$2,'operations','2026-10-01','2026-10-31',60000000)", "BUDGET-"+line, line)
	}
	results := tx.SendBatch(ctx, batch)
	if e = results.Close(); e != nil {
		return false, e
	}
	if e = tx.Commit(ctx); e != nil {
		return false, e
	}
	return true, nil
}
