# Roadmap Implementasi

Arsitektur disetujui melalui instruksi implementasi pengguna. Kapasitas 3–5 jam/hari; estimasi sekitar 36 jam, bukan jaminan seluruh backlog selesai.

| Tanggal 2026 | Hasil |
|---|---|
| 27 Oktober | Dokumen, Go/Docker, fondasi, skema dan konektivitas Meta |
| 28 Oktober | Migrasi, seed 60 hari dan validasi golden |
| 29 Oktober | 4 aturan, fakta/bukti, API kasus |
| 30 Oktober | Agent, 8 tool, checkpoint, budget, laporan |
| 31 Oktober | Golden end-to-end melalui UI sampai review |
| 1 November | Error, missing data, cancel, retry, restart |
| 2 November | Login, deploy gratis, SSE dan verifikasi daring |
| 3 November | Latihan demo, README, diagram, reproduksi |
| 4 November | Freeze, rekaman, rules final, submit 18.00 WIB |

Urutan dependensi: fondasi→schema→Meta adapter→seed→deteksi→agent→golden→UI→reliabilitas→deploy/demo. Konektivitas diperiksa sedini mungkin untuk menghindari adapter palsu yang baru ditemukan pada akhir. Pipeline lokal dan CI memakai mock eksplisit; golden acceptance nyata memakai Meta.

Jika 31 Oktober terancam: tunda ATT-001/OT-002/COST-001/COST-002, filter tambahan, visual polishing, report export. Pertahankan model nyata, semua bukti golden, login, review, audit, queue recovery dan budget. Jangan menambah ERP, mikroservis, Kafka, Kubernetes, Python service, multitenancy atau prediksi.

Risiko prioritas: runtime/toolchain belum siap; key/kredit tidak tersedia di environment; kelengkapan sumber disalahartikan sebagai pelanggaran; hallucination meski JSON valid; cold start hosting; biaya retry saat crash; aturan kompetisi belum lengkap. Mitigasi masing-masing ada di ARCHITECTURE, AI_AGENT_DESIGN dan HACKATHON_DEMO_PLAN.
