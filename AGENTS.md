# Petunjuk proyek OpsMind AI

Sebelum mengubah proyek, baca docs/HANDOFF.md, docs/ARCHITECTURE.md dan docs/IMPLEMENTATION_ROADMAP.md. Handoff mencatat hasil nyata, kendala, status jeda, dan langkah berikutnya. Lanjutkan hanya atas instruksi pengguna.

## Cara bekerja
- Gunakan bahasa Indonesia. Berikan penilaian langsung, analitis dan faktual; jelaskan asumsi lemah, risiko, serta pekerjaan yang benar-benar belum selesai.
- Arsitektur telah disetujui: monolit modular Go, PostgreSQL, React/TypeScript, Meta Model API. Jangan meminta persetujuan arsitektur ulang.
- P0 hanya OT-001, PROD-001, ATT-002, MAINT-001 dan golden investigation lengkap. Pengguna mengaktifkan pengerjaan P1 pada 28 Oktober 2026: ATT-001, OT-002, COST-001, COST-002. Modul lanjutan lain tetap ditunda. Demo dan video ditangani pengguna sendiri.
- Jangan menambah microservices, Kafka, Kubernetes, Python service, ERP, multi-fasilitas, registrasi publik atau fitur di luar plan.
- Semua data manufaktur sintetis. Empat attendance pending bukan bukti mangkir; cuti disetujui harus dikecualikan. Korelasi bukan pembuktian sebab.
- Fakta dan uang dihitung Go; model hanya melalui tool baca yang diizinkan. Review mencatat keputusan tanpa menulis data bisnis.
- Identitas author dan committer hanya akun pengguna Tegar234-koplox (Tegar Sang Putra). Jangan menambahkan GPT, bot, atau trailer Co-authored-by otomatis.
- Jangan mencetak, commit atau mengunggah .env, API key, password, session secret, runtime .local/.tools, trace browser, node_modules atau binari.
- Anggaran tunai US$0. Jangan membuat layanan berbayar. Batas kredit aplikasi US$20 (15 development/5 demo), maksimum US$0,25/run.
- Jangan mengklaim deployment, CI remote, publikasi, atau submission berhasil hanya karena config tersedia.

## Pengujian
Gunakan TEST_DATABASE_URL dengan nama database berakhiran _test; integration test mereset schema database tersebut. Jangan jalankan reset integration bersamaan dengan server yang memakai database _test dan worker lock yang sama.
Mock dipilih eksplisit, selalu berlabel. Tes Meta nyata terpisah dan tidak menjadi syarat CI. Jalankan pemeriksaan sesuai perubahan; jangan mengulang seluruh pengujian yang sudah lulus tanpa alasan baru.

## Alur Git
- Untuk perubahan berikutnya gunakan branch codex/, commit atas nama akun pengguna, push branch, lalu buat PR ke main.
- Lakukan squash and merge ketika PR siap dan pemeriksaan yang relevan lulus. Jangan push perubahan berikutnya langsung ke main.
- Author dan committer tetap akun pengguna; jangan menambahkan GPT, bot, atau trailer Co-authored-by otomatis.
