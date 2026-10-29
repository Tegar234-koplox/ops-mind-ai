# Desain Agent Investigasi

## Model dan kontribusi
Adapter HTTP Go ke https://api.meta.ai/v1/chat/completions, model muse-spark-1.3 Standard, kredensial MODEL_API_KEY hanya backend. Model memilih bukti tambahan dan membandingkan kemungkinan penjelasan. Go menghitung angka, memeriksa scope dan merekonstruksi observasi terverifikasi. Tool calling dan JSON schema tersedia menurut https://dev.meta.ai/docs/tool-calling dan https://dev.meta.ai/docs/structured-output .

## Kontrak tool
Allowlist: get_shift_staffing, get_attendance_summary, get_approved_leave, get_overtime_variance, get_cost_breakdown, get_production_performance, get_downtime_events, get_evidence_record. Semua menerima shift_id yang harus identik dengan scope kasus; get_evidence_record juga evidence_id milik kasus. additionalProperties=false; parameter lain ditolak. Tool mengembalikan evidence_id, facts terstruktur, SourceRef, completeness, serta catatan data yang kurang. SQL tetap di backend.

Tool mengorelasikan shift_instances sebagai konteks tunggal; assignment menghubungkan employee/cuti/punch/lembur; downtime menghubungkan mesin/lini/overlap waktu; produksi dan biaya menggunakan lini/shift sama. Tidak meminta nama atau data kesehatan. Evidence disimpan sebelum diberikan ke model dan dapat dibuka dari UI.

## State dan checkpoint
queued → running → succeeded | insufficient_evidence | failed | cancelled. Transcript disimpan setelah respons model dan tiap tool. Saat resume, tool call yang belum memiliki output dieksekusi; output yang sudah tersimpan digunakan kembali. Setiap tool menggunakan cache run + call_id + hash argumen. Resume tidak mengulang tool yang selesai. Budget request yang hasilnya hilang tetap ditagih pada reservasi konservatif.

Maksimum 8 putaran, 16 tool, 180 detik; timeout model 30 detik dan tool 5 detik; maksimum 2 retry sementara total, exponential backoff terbatas dan Retry-After dihormati jika muat sisa waktu. 400/401/403 bukan retry. Satu worker aktif dilindungi DB lock. Tidak ada built-in web search, shell, SQL atau tool tulis.

## Laporan
summary; observation_evidence_ids; hypotheses (statement, evidence_ids, confidence low/medium, counter_evidence); information_gaps; recommendations (title, rationale, evidence_ids, department). JSON schema ketat serta validasi server. Golden investigation wajib memiliki bukti workforce, biaya, produksi dan maintenance. Angka fakta UI dibaca dari evidence Go; referensi asing, confidence kausal pasti, dan klaim misconduct/fraud ditolak. Validasi teknis tidak membuktikan kebenaran seluruh teks: narasi tetap hipotesis dan memerlukan peninjau. Keluaran tidak lengkap ditandai insufficient_evidence dan tidak menerbitkan rekomendasi yang dapat disetujui.

## Biaya dan mode
Total aplikasi US$20: development 15 dan demo 5, maksimum run 0,25. Reservoir biaya persisten memakai transaksi serialisasi DB; sebelum panggilan reserve harga maksimum input konservatif dan max_completion_tokens, setelah respons rekonsiliasi usage. Reservasi request yang putus dipertahankan sebagai pengeluaran maksimum. Tarif konfigurasi awal input US$1,25 / juta dan output US$4,25 / juta (cached input dihitung tarif penuh demi konservatif), per https://dev.meta.ai/docs/pricing-rate-limits . Tidak auto top-up atau beralih model. Mode mock tersedia eksplisit untuk CI, memakai tool loop dan diberi label SIMULASI di UI; tidak pernah fallback diam-diam dari Meta.

## Bukti aturan P1
Delapan nama tool tetap sama. get_attendance_summary menyertakan rangkaian absen terkonfirmasi untuk ATT-001; get_overtime_variance menyertakan status persetujuan efektif untuk OT-002; get_cost_breakdown menyertakan baseline/variance/closure untuk COST-001/COST-002. Referensi historis hanya diizinkan jika dipilih query deterministik untuk kasus tersebut; kasus lain tetap pada scope awal. Snapshot dideduplikasi, di-hash, dan tidak memberi akses SQL/model tulis.
# P2 serah terima tidak menambah agent

Builder laporan shift memakai fakta deterministik Go dan membaca sintesis/rekomendasi dari investigasi yang sudah tersimpan. Mode mock pada sintesis tetap berlabel SIMULASI. Tidak ada panggilan Meta, reservasi kredit, tool AI baru, perubahan loop, atau penulisan data bisnis melalui laporan. Delapan tool dan semua batas anggaran tetap sama. Laporan bukan investigasi baru; bila perlu bukti tambahan, operator membuka kasus dan memulai alur investigasi yang sudah tersedia.
