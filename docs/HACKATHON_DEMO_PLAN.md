# Rencana Demo dan Submission

## Naskah 180 detik
0–20: manajer menghadapi biaya lembur naik dan produksi turun; buka kasus OT-001. 20–45: tunjukkan Go menghitung +38% dari baseline yang sebanding. 45–100: mulai/ikuti agent Meta meminta lembur, staffing, cuti, produksi dan downtime; timeline menampilkan status tool, bukan chain-of-thought. 100–140: buka bukti 31 hadir +3 cuti +4 belum terverifikasi, output 830/1000, downtime union 110 menit. 140–165: tampilkan hipotesis, keterbatasan kausal dan rekomendasi. 165–180: review satu rekomendasi, tunjukkan audit dan tanpa perubahan otomatis pada data bisnis.

## Persiapan
MODEL_MODE=meta, key tersedia hanya backend, fase budget demo dan sisa kredit cukup. Dataset seed 20261027; login penilai dibagikan melalui kanal submission yang aman, bukan repository/video. Buka deployment sebelum rekaman untuk cold start. Siapkan Docker Compose sebagai cadangan dan simpan hasil investigasi nyata. Laporan tersimpan/rekaman ulang diberi label; mock tidak pernah diklaim panggilan Meta nyata.

## Pengiriman
Public GitHub: README, setup, .env.example tanpa secret, 10 dokumen, ERD dan diagram arsitektur, migrasi, seed, pengujian, license/attribution. Video 180 detik adalah target proyek; cek durasi, bahasa dan format yang benar di portal. Lampirkan URL terlindungi, instruksi login dan batas cold start. Jangan mengunggah key atau kredensial penilai ke public repo. Sumber event https://dev.meta.ai/events/global-hackathon; halaman publik mengonfirmasi model, track dan deadline, tetapi rules final belum terakses saat perencanaan. Verifikasi portal sebelum submit, termasuk syarat karya baru dan penggunaan coding assistant.

Deadline resmi: 4 November 2026 21.00 ET =5 November 09.00 WIB (America/New_York saat itu EST). Internal: 4 November 18.00 WIB. Bangun hanya dengan data dan aset sintetis; tidak mengklaim materi sebelum periode event sebagai karya baru. Rekaman dan submission merupakan tindakan terpisah yang harus benar-benar selesai sebelum dinyatakan terkirim.

## Sumber persyaratan dan penanggung jawab

Ringkasan aturan/onboarding peserta menetapkan kategori Indie Developer, pekerjaan baru 26 Oktober–4 November 2026, repository publik, video tiga menit dan submission melalui Cerebral Valley. Status tiap sumber serta checklist sebelum kirim ada di [SUBMISSION_REQUIREMENTS](SUBMISSION_REQUIREMENTS.md); bukti fitur di [RELEASE_READINESS](RELEASE_READINESS.md). Codex digunakan sebagai coding assistant, tetapi belum ada bukti persetujuan eksplisit panitia. Jangan mengubah ketiadaan larangan yang ditemukan menjadi klaim persetujuan.

Demo, video dan submission ditangani pengguna sendiri. Agent tidak melanjutkan rekaman atau mengirim formulir tanpa instruksi. Tetap prioritaskan alur golden; fitur P1 dan serah terima hanya ditampilkan bila sesuai waktu dan benar-benar diperagakan. Tidak perlu menambah fitur untuk mengisi video.
