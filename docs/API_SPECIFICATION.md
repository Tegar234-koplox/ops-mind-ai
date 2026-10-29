# Spesifikasi API v1

Origin tunggal, JSON, prefix /api/v1. Login POST /auth/login (username, password); GET /auth/session; POST /auth/logout. Cookie sesi signed, HttpOnly, SameSite Strict, Secure pada HTTPS. Endpoint selain login dan /healthz memerlukan login. Mutasi memerlukan Origin yang cocok serta Idempotency-Key 8–128 karakter (login/logout dikecualikan). Error: {error:{code,message},request_id}; tidak mengungkap token atau SQL.

| Metode | Endpoint | Hasil |
|---|---|---|
| GET | /overview | Ringkasan kasus, mode model, kuota, seed/as-of sintetis |
| POST | /detections | Jalankan empat aturan P0 pada shift selesai; optional shift_id; enqueue OT-001 otomatis |
| GET | /cases?limit=25&offset=0&status=&severity= | {items,total}; limit maksimum 100 |
| GET | /cases/{id} | Kasus, investigasi terbaru, rekomendasi dan review |
| POST | /cases/{id}/investigations | 202 pekerjaan, {id,status}; optional instruction |
| POST | /cases/{id}/follow-up-investigations | 202 investigasi baru, instruction wajib |
| GET | /investigations/{id} | Status, progres, laporan, evidence dan event terbaru |
| GET | /investigations/{id}/events | text/event-stream; Last-Event-ID atau after; kejadian persisten berurutan |
| POST | /investigations/{id}/cancel | Pembatalan persisten dan audit |
| GET | /cases/{id}/evidence/{evidenceId} | Snapshot, hash, sumber, fakta dan completeness |
| POST | /recommendations/{id}/reviews | {decision:approved|rejected,reason}; rejected wajib reason |

case: id, rule_id, entity_id, shift_id, department, severity, status, data_quality, expected, observed, explanation, detected_at. Nilai uang {currency:IDR,amount:string}; persentase integer bps. run: queued, running, succeeded, insufficient_evidence, failed, cancelled. Case open → investigating → review_required → reviewed; failed dan insufficient_evidence kembali open. Follow-up tidak menghapus audit lama; rekomendasi lama tidak dapat direview setelah run baru dimulai.

POST idempotensi terikat pengguna, path dan hash body. Kunci sama/body berbeda →409; proses belum selesai →409 dengan Retry-After; respons selesai diputar ulang. Rekomendasi hanya satu keputusan, run terbaru harus succeeded. Persetujuan tidak mengeksekusi tindakan bisnis. Status 400 validation, 401 unauthenticated, 403 origin/scope, 404 record, 409 conflict, 429 rate/budget, 503 dependency unavailable. Semua error menyertakan request_id.

SSE jenis: queued, stage, tool_started, evidence_collected, model_response, completed, failed, cancelled. Heartbeat komentar 10 detik; koneksi ditutup setelah status terminal dan event terkirim. Resume hanya event ID > cursor; frontend deduplicate ID dan polling 3 detik saat stream gagal. HTTP disconnect tidak membatalkan worker. API JSON timeout terbatas; SSE dikecualikan dari write timeout pendek. Login 5 upaya/menit/IP, mutasi 30/menit/operator; satu run aktif global.

## Metadata P1
GET /overview menambahkan active_rule_count: 4 atau 8 sesuai ENABLE_P1_RULES. Tidak ada endpoint untuk mengubah kehadiran, approval, closure posting atau menjalankan seed. Fixture P1 hanya CLI/flag environment eksplisit. Kontrak delapan tool tetap; fakta P1 dan referensi histori ditambahkan pada snapshot untuk aturan terkait. Case kehadiran diagregasi per lini/shift, bukan keputusan pelanggaran individu.
# P2: laporan serah terima shift

Semua endpoint berikut memerlukan sesi operator. Scope tetap `FACTORY-01`; ID fasilitas lain ditolak atau tidak ditampilkan. Penyusunan tidak memanggil model atau menulis data bisnis.

| Metode dan rute `/api/v1` | Kontrak |
|---|---|
| `GET /shifts` | Shift berakhir paling lambat waktu evaluasi fixture. Filter `line_id`, `date=YYYY-MM-DD`, pagination `limit` 1–100/default 30 dan `offset` >=0. Respons `items`, `total`, `lines`, `as_of`. Tanggal operasional mengikuti awal shift Asia/Bangkok. |
| `POST /handovers` | Body `{ "shift_id": "A-20261026-S1" }`, `Idempotency-Key` wajib. HTTP 201 dengan laporan lengkap; replay mengembalikan byte respons yang sama. Key sama/payload berbeda 409. Shift belum selesai 409 `shift_not_finished`, tidak ditemukan/di luar fasilitas 404, payload invalid 400. |
| `GET /handovers` | Riwayat ringkas: `id`, `shift_id`, `created_at`, `created_by`; filter opsional `shift_id`, pagination `limit`/`offset` seperti di atas. Respons `items`, `total`. |
| `GET /handovers/{id}` | Snapshot lengkap yang immutable, hash SHA-256, pembuat dan waktu pembuatan. Tidak ada PUT/DELETE. |

Snapshot versi 1 memuat `facility_id`, `timezone`, `synthetic`, `data_as_of`, `shift`, `next_shift` atau null, empat `sections` berisi facts string/quality/notes, kasus pada shift beserta latest run dan rekomendasi/review-nya, interval gangguan, referensi sumber bertipe, dan catatan keterbatasan. Rupiah dikirim sebagai string. `created_at` adalah waktu snapshot; `data_as_of` adalah cutoff evaluasi data operasional sintetis. Keputusan review dan investigasi yang sudah tersimpan dibaca saat penyusunan, bukan direkonstruksi seolah-olah ada pada cutoff operasional.

`hash` dihitung dari snapshot JSON bertipe setelah canonical marshaling Go, bukan byte JSONB mentah. Hash memeriksa konsistensi snapshot, tidak membuktikan kebenaran seluruh narasi. Penyusunan versi baru tidak menimpa versi lama. Cetak/PDF dilakukan melalui browser dari snapshot yang sedang dibuka.
