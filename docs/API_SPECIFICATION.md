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
