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

## Perubahan urutan atas instruksi pemilik, 28 Oktober
P0 dan golden daring sudah stabil; pengguna menangani demo/video sendiri dan meminta pengerjaan ATT-001, OT-002, COST-001, COST-002 terlebih dahulu. P1 tetap dalam monolit dan delapan tool baca. Tidak memperluas ke ERP atau modul lanjutan. Syarat penyelesaian: migrasi upgrade aman, fixture terpisah, tes batas/sumber tidak lengkap, browser, CI, PR dan deployment terverifikasi.

## P2 dipilih pengguna: serah terima shift

P0/P1 dan hover sudah selesai; pengguna memilih laporan serah terima shift dari Phase 2 brief. Cakupan hanya laporan atas permintaan untuk shift selesai, snapshot immutable, fakta empat domain dari Go, kasus/sintesis/review yang sudah tersimpan, riwayat versi dan cetak browser. Tidak menambah detektor, panggilan Meta, scheduler, notifikasi, penerima atau tindakan bisnis otomatis.

Urutan: kontrak snapshot/API → migrasi 000005 dan pembacaan konsisten → builder Go → workspace terpisah → tes ketepatan/data kurang/versi/idempotensi/browser/cetak → commit/push/PR/CI → squash merge sesuai aturan identitas → deploy gratis dan verifikasi tanpa tambahan biaya model. Domain Phase 2 lainnya tetap ditunda. Kriteria selesai mencakup golden 31+3+4, +38%, -17%, 110 menit; snapshot lama tidak berubah setelah review baru.

## Kesiapan rilis, 29 Oktober

P0, P1, hover dan P2 sudah selesai serta diverifikasi daring melalui PR #1–#4. Tahap berikutnya yang diminta pengguna adalah kesiapan rilis: lengkapi pengujian kegagalan provider yang belum langsung diuji, selaraskan README/handoff dengan hasil nyata, simpan atribusi dependensi dan pisahkan persyaratan onboarding dari aturan publik yang berhasil diverifikasi. Tidak ada modul bisnis baru.

Urutan: pengujian terarah lokal tanpa kredit Meta → dokumen bukti dan checklist → commit/push/PR/CI → squash merge sesuai batas izin identitas → verifikasi paket/deployment bila kemasan runtime berubah. Daftar bukti berada di RELEASE_READINESS.md; ketentuan acara dan hal yang masih harus dicek berada di SUBMISSION_REQUIREMENTS.md. Demo, rekaman tiga menit, akses juri dan pengiriman final ditangani pengguna. Jadwal asli tetap menjadi batas; pekerjaan selesai lebih awal tidak membuka scope baru.
