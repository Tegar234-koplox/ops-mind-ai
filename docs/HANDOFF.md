# Handoff OpsMind AI

Diperbarui 28 Oktober 2026 (WIB). Status: pekerjaan dilanjutkan atas instruksi pengguna; deployment dan golden daring terverifikasi.

## Keputusan tetap
- Monolit modular Go/PostgreSQL/React/Meta Model API; hanya OT-001, PROD-001, ATT-002, MAINT-001 dan golden investigation.
- P1, ERP, microservices, multi-fasilitas, prediksi dan registrasi publik ditunda.
- Seluruh data manufaktur sintetis. Empat attendance pending bukan bukti mangkir; tiga cuti disetujui dikecualikan. Korelasi bukan kausalitas. Review tidak menulis data bisnis.
- Tunai US$0. Batas aplikasi US$20 (15 development/5 demo), US$0,25/run. Tidak membeli layanan atau domain.
- Author dan committer hanya Tegar Sang Putra / Tegar234-koplox, tanpa GPT/bot atau Co-authored-by.
- Git: branch codex/ → commit → push branch → PR → squash and merge ketika siap dan checks lulus. Jangan push perubahan berikutnya langsung ke main.

## Deployment yang benar-benar sudah berjalan
- URL: https://opsmind-ai-l0px.onrender.com ; login wajib, kredensial sama dengan operator lokal dan dibagikan privat oleh pemilik.
- Render service srv-dbh00g6k1f9s7399lt60, workspace tea-dbgsfrh42hec73du46rg, Docker plan free, region singapore, branch main, autoDeployTrigger checksPass.
- Neon project wild-dew-57741693 (opsmind-ai), PostgreSQL 18, region aws-ap-southeast-1, compute min/max 0,25 CU, auto-suspend default paket Free. Organisasi diverifikasi plan free sebelum pembuatan.
- Endpoint ep-broad-mouse-b35rsj8q; database opsmind. Gunakan koneksi langsung Neon dengan sslmode=verify-full karena worker menggunakan session advisory lock.
- Dua proyek Neon lain milik pengguna tidak diubah. Layanan Key Value yang pernah dibuat pengguna bukan dependensi aplikasi dan tidak diubah.
- Commit aplikasi yang terpasang: 3d6ef8aca859df8b6d37b570335caba43c920d13 (main). PR readiness hanya mengubah CI, petunjuk dan konfigurasi deployment; belum ada perubahan runtime aplikasi setelah commit ini.
- Build pertama dep-dbh00h6k1f9s7399lv8g berhasil. Karena URL final baru diketahui setelah create, APP_ORIGIN diperbarui lalu image yang sama diterapkan ulang (dep-dbh01hu1egvs73e9pj60) agar login memakai origin final.
- Restart terkontrol berikutnya dep-dbh03n6gekts73904ne0 berstatus live; laporan/bukti/review/biaya/seed tetap utuh.
- Environment backend memakai MODEL_MODE=meta, muse-spark-1.3, BUDGET_PHASE=demo, COOKIE_SECURE=true, password bcrypt, session secret acak, SEED_IF_EMPTY=true. Tidak ada management key pada backend.
- Render Free dapat tidur saat idle. Tidak ada jaminan worker berjalan 24/7; cold start dan kuota gratis tetap berlaku.

## Bukti verifikasi daring
- Health/frontend HTTP 200; HTML bahasa Indonesia; API tanpa login 401; origin asing 403; login benar 200 dengan cookie Secure/HttpOnly/SameSite=Strict.
- Database baru berisi 150 pekerja, 3 lini, 8 mesin dan 60 hari dari seed 20261027 terbaru. Tarif lembur generator Rp500/menit konsisten; tidak memakai database lokal lama.
- Deteksi golden A-20261026-S1 menghasilkan empat aturan P0. Replay Idempotency-Key identik dan tidak membuat investigasi kedua.
- Case 6cf2a0b1f6e4a08521ec6619cb960c2c; run fc4ed857586151f2e08aed563cbf4dbe succeeded, Meta nyata, 2 putaran, 7 tool, 7 evidence, 4 domain, 2 hipotesis, 3 rekomendasi.
- Angka golden: baseline Rp1.000.000; aktual Rp1.380.000 (+38%); output 830/1000 (-17%); penugasan 38 =31 hadir terverifikasi +3 cuti +4 pending; downtime union 110 menit.
- Laporan mempertahankan ketidakpastian empat pending, mengakui cuti sah, dan tidak menyatakan korelasi sebagai kausalitas.
- SSE pertama 9 kejadian, reconnect 4 kejadian lebih baru tanpa duplikasi. Bukti lintas kasus ditolak 404; semua hash snapshot valid secara format dan evidence scoped ke run/kasus.
- Playwright khusus deployment menjalankan meta.spec.ts: 1 passed (16,2 detik). Login → laporan Meta → bukti downtime → review → audit berhasil.
- Satu rekomendasi approved dengan catatan uji data sintetis, dua pending. Audit mencatat operator; tidak menjalankan perubahan bisnis.
- Setelah restart: run sama succeeded, evidence 7, approved 1/pending 2, pekerja 150 dan biaya tidak berubah.
- Biaya run daring US$0,018857; saldo demo database daring US$0,018857. Checkpoint biaya lokal sebelumnya US$0,059159 (termasuk reservasi konservatif kegagalan jaringan); gabungan tercatat US$0,078016. Counter dua database terpisah, jangan menjumlahkan dua limit menjadi anggaran baru.
- Blueprint render.yaml diperbarui Singapore/main/checksPass dan divalidasi API Render: HTTP 200, valid=true, tanpa provisioning tambahan.

## Git dan CI
- Repo publik: https://github.com/Tegar234-koplox/ops-mind-ai
- PR #1: https://github.com/Tegar234-koplox/ops-mind-ai/pull/1 ; sudah dilampirkan pada chat, masih terbuka dan belum squash merge.
- Branch codex/deployment-readiness; commit readiness awal 13b29d84a7dd956007b6647ca72ae841e01dfe68 memakai author dan committer akun pengguna.
- CI main awal lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40602651746
- CI push readiness lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40604969244
- CI PR readiness lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40605232154
- Job verify mencakup vet/race test/PostgreSQL 18/frontend/Playwright mock; job Container startup membangun Dockerfile dan memverifikasi Compose/login/seed. Docker lokal belum menyediakan daemon; validasi container nyata telah lulus di runner Linux.
- README, Blueprint dan handoff hasil deployment disiapkan sebagai commit lanjutan pada PR yang sama. Periksa checks untuk SHA head terbaru; jangan menyamakan hasil CI commit lama dengan head baru.
- Syarat committer hanya pengguna berpotensi konflik dengan squash GitHub yang mencatat noreply@github.com. REST merge tidak menyediakan committer kustom. Jangan merge atau push langsung ke main untuk mengakali aturan; klarifikasi pengecualian metadata teknis bila diperlukan. Author/contributor GPT atau bot tetap dilarang.
- Transport Git yang berhasil tanpa perubahan global: git -c http.sslBackend=schannel -c http.curloptResolve=github.com:443:20.205.243.166 -c credential.helper= -c 'credential.helper=!gh auth git-credential' ... ; resolve IPv4 ulang jika alamat berubah.
- GitHub CLI login akun Tegar234-koplox; konektor write sebelumnya 403. PR dibuat via CLI akun pengguna, tanpa mencetak token.

## Rahasia dan artefak lokal
- .env: password operator, session secret lokal, Meta API key. .env.deploy: NEON_API_KEY dan RENDER_API_KEY. Keduanya valid saat provisioning; jangan cetak/commit/upload.
- .local/cloud-state.json menyimpan DSN Neon rahasia dan ID resource; .local/cloud-report.json serta cloud-verification.json menyimpan bukti tes. Semuanya di-ignore Git dan Docker.
- Helper API hanya lokal di .local/, memakai HTTPS IPv4 dan timeout, tidak me-retry mutasi otomatis. Respons credential disimpan lokal tanpa dicetak.
- Screenshot daring di web/test-results/cloud-meta-report-desktop.png dan cloud-meta-review-desktop.png; sudah diperiksa visual dan di-ignore. Tidak ada trace browser yang diunggah.
- Runtime lokal: Go 1.27.3 .tools/go; Node 24; PostgreSQL portabel port 55432. Server Meta lokal lama mungkin masih port 8080; cek proses sebelum menjalankan ulang.
- opsmind_test khusus tes reset schema; jangan jalankan reset bersama server/worker pada DB test yang sama.
- Ground truth eval/ tidak dibaca runtime/model. Validasi schema/hash/referensi tidak membuktikan kebenaran semua narasi.

## Urutan lanjut yang dibatasi plan
1. Selesaikan commit/push README/Blueprint/handoff pada PR #1, periksa CI SHA terbaru, dan selesaikan keputusan metadata committer sebelum squash merge.
2. Setelah merge, pastikan main/Render menunjuk revisi yang benar dan golden tersimpan tetap dapat dibuka; tidak perlu mengulang run Meta tanpa alasan.
3. Siapkan video 180 detik, akses juri privat, dan cadangan rekaman investigasi tersimpan dengan label jelas. Video dan submission belum dilakukan.
4. Cocokkan aturan final di portal peserta. Target internal 4 November 2026 18.00 WIB; tenggat event menurut plan 5 November 09.00 WIB.
5. Jangan menambah aturan P1 atau modul di luar plan tanpa keputusan pengguna.
