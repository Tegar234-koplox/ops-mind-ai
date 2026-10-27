# Desain Database

PostgreSQL 18. PK teks deterministik untuk data sintetis dan UUID acak aplikasi. FK mencegah orphan; operasi bisnis memakai transaksi. BIGINT untuk IDR, timestamptz untuk instan, date untuk tanggal operasional. JSONB hanya untuk transcript, hasil terstruktur dan snapshot, bukan pengganti relasi operasional.

```mermaid
erDiagram
 departments ||--o{ employees : employs
 departments ||--o{ production_lines : owns
 production_lines ||--o{ machines : contains
 production_lines ||--o{ shift_instances : operates
 calendar_days ||--o{ shift_instances : schedules
 shift_instances ||--o{ shift_assignments : assigns
 employees ||--o{ shift_assignments : works
 shift_assignments ||--o| attendance_records : records
 employees ||--o{ leave_records : requests
 shift_assignments ||--o{ overtime_records : accrues
 shift_instances ||--o| production_runs : produces
 machines ||--o{ downtime_events : interrupts
 shift_instances ||--o{ downtime_events : experiences
 production_lines ||--o{ expenses : spends
 production_lines ||--o{ budgets : plans
 anomaly_cases ||--o{ investigation_runs : investigates
 shift_instances ||--o{ anomaly_cases : triggers
 investigation_runs ||--o{ case_evidence : retrieves
 investigation_runs ||--o{ tool_executions : executes
 investigation_runs ||--o{ investigation_events : streams
 investigation_runs ||--o{ recommendations : proposes
 recommendations ||--o| review_actions : decides
 anomaly_cases ||--o{ audit_events : audits
```

## Tabel dan integritas
16 tabel wajib: departments, employees, production_lines, shift_assignments, attendance_records, leave_records, overtime_records, production_runs, downtime_events, expenses, budgets, anomaly_cases, investigation_runs, case_evidence, review_actions, audit_events. Tambahan: machines, calendar_days, shift_instances, investigation_events, tool_executions, recommendations. Tabel teknis api_idempotency menyimpan replay respons dan budget_reservations menyimpan biaya maksimum/aktual per panggilan.

Shift dan interval memiliki CHECK end > start. Satu assignment per employee/shift; query seed dan uji memastikan tidak ada shift pekerja overlap. Satu attendance per assignment. Observed output, target, biaya dan durasi tidak negatif. Punch tidak lengkap diperbolehkan untuk menguji missing data; keduanya terisi harus berurutan. Relasi mesin/lini dan assignment/employee dicek backend serta seed. Deteksi unik (rule_id, entity_id, period_start, rule_version); hanya satu run aktif per kasus. Review unik per rekomendasi dan hanya berlaku pada laporan terbaru yang berhasil divalidasi.

case_evidence menyimpan snapshot kanonik, hash SHA-256, daftar SourceRef bertipe, case_id/run_id dan scope shift. Referensi polimorfik sumber divalidasi backend melalui allowlist tabel dan ID, bukan diklaim sebagai FK lintas tabel. Evidence, recommendation dan review memiliki FK nyata ke kasus/run. Indeks penting: shift(line_id, starts_at); assignment(employee_id, shift_id); run(status, lease_until); event(run_id, id); evidence(case_id, run_id); expense(line_id, occurred_at); audit(case_id, created_at).

## Migrasi
000001: data operasional dan constraint. 000002: kasus, antrean, event, tool, evidence, rekomendasi, review, audit, idempotensi dan budget. Pasangan up/down; rollback menghapus dari dependensi paling bawah. Migrate memakai advisory lock. Seed Go terpisah, transactional, reproducible dan berhenti jika dataset sudah terisi. Tidak mereset database melalui endpoint. Tidak menyimpan ground truth di database aplikasi.
