# Handoff OpsMind AI

Status terbaru: P0/P1 dan hover selesai serta terverifikasi daring; pengguna memilih P2 laporan serah terima shift. Baca .local/LAST_PROGRESS.md untuk checkpoint verifikasi sesudah commit.

## Keputusan tetap
- Monolit modular Go/PostgreSQL/React/Meta Model API; P0 OT-001, PROD-001, ATT-002, MAINT-001 dan golden tetap dipertahankan. P1 hanya empat aturan yang diminta.
- Pengguna mengaktifkan ATT-001, OT-002, COST-001, COST-002 pada 28 Oktober dan meminta clean code. ERP, microservices, multi-fasilitas, prediksi dan registrasi publik tetap ditunda.
- P2 diizinkan hanya laporan serah terima shift. Pakai snapshot fakta Go dan hasil investigasi/review tersimpan; tanpa model/tool tambahan, scheduler, notifikasi, penerima atau tindakan bisnis otomatis. Domain Phase 2 lain tetap ditunda.
- Demo/video ditangani pengguna sendiri. Jangan melanjutkan persiapan video. Petunjuk/kredensial lokal ada di .local/DEMO_ACCESS.md (privat, di-ignore, jangan cetak atau upload).
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
- Commit yang terpasang pada checkpoint awal P2: 9686ceaffe6c446ed9c9ff48cb18fceef167d09e (hasil squash PR #3 pada main). P1 dan hover aktif; P2 belum dipasang sampai verifikasi rilisnya dicatat.
- Build pertama dep-dbh00h6k1f9s7399lv8g berhasil. Karena URL final baru diketahui setelah create, APP_ORIGIN diperbarui lalu image yang sama diterapkan ulang (dep-dbh01hu1egvs73e9pj60) agar login memakai origin final.
- Restart terkontrol dep-dbh03n6gekts73904ne0 mempertahankan laporan/bukti/review/biaya/seed. Deployment P0 dep-dbh096ei0phs73ad1280 sebelumnya live pada c3fd411; digantikan deployment P1 di bawah.
- Deployment final dipicu API setelah CI main lulus. autoDeployTrigger=checksPass terkonfigurasi, tetapi pemicuan otomatis belum terverifikasi; jangan mengklaim hook CI-to-Render otomatis berhasil.
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
- Setelah restart dan setelah deployment hasil merge: health/login berhasil, run sama succeeded, evidence 7, approved 1/pending 2, pekerja 150 dan biaya tidak berubah. Tidak ada panggilan Meta tambahan.
- Biaya run daring US$0,018857; saldo demo database daring US$0,018857. Checkpoint biaya lokal sebelumnya US$0,059159 (termasuk reservasi konservatif kegagalan jaringan); gabungan tercatat US$0,078016. Counter dua database terpisah, jangan menjumlahkan dua limit menjadi anggaran baru.
- Blueprint render.yaml diperbarui Singapore/main/checksPass dan divalidasi API Render: HTTP 200, valid=true, tanpa provisioning tambahan.

## Git dan CI
- Repo publik: https://github.com/Tegar234-koplox/ops-mind-ai
- PR #1: https://github.com/Tegar234-koplox/ops-mind-ai/pull/1 ; squash merge BERHASIL menjadi c3fd4118721c8b6a6608b62a3a1c240e8d5cbf03. Author Tegar Sang Putra; committer teknis GitHub noreply@github.com diizinkan pengguna khusus PR #1. Tidak ada Co-authored-by. Daftar contributors API hanya Tegar234-koplox.
- Branch codex/deployment-readiness; commit readiness awal 13b29d84a7dd956007b6647ca72ae841e01dfe68 memakai author dan committer akun pengguna.
- CI main awal lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40602651746
- CI push readiness lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40604969244
- CI PR readiness lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40605232154
- Job verify mencakup vet/race test/PostgreSQL 18/frontend/Playwright mock; job Container startup membangun Dockerfile dan memverifikasi Compose/login/seed. Docker lokal belum menyediakan daemon; validasi container nyata telah lulus di runner Linux.
- Commit akhir PR 2660dedc0390e381c6107d91f57670eca1cd60c6 lulus CI: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40703708067
- CI main hasil squash c3fd411 LULUS (verify dan Container startup): https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40704406798
- Pengecualian committer GitHub diberikan eksplisit untuk PR #1 dan, melalui jawaban terpisah, hanya PR #2. Bukan izin umum untuk PR berikutnya; aturan AGENTS.md tetap berlaku. Author/contributor GPT atau bot tetap dilarang.
- Transport Git yang berhasil tanpa perubahan global: git -c http.sslBackend=schannel -c http.curloptResolve=github.com:443:20.205.243.166 -c credential.helper= -c 'credential.helper=!gh auth git-credential' ... ; resolve IPv4 ulang jika alamat berubah.
- GitHub CLI login akun Tegar234-koplox; konektor write sebelumnya 403. PR dibuat via CLI akun pengguna, tanpa mencetak token.

## Rahasia dan artefak lokal
- .env: password operator, session secret lokal, Meta API key. .env.deploy: NEON_API_KEY dan RENDER_API_KEY. Keduanya valid saat provisioning; jangan cetak/commit/upload.
- .local/cloud-state.json menyimpan DSN Neon rahasia dan ID resource; .local/cloud-report.json serta cloud-verification.json menyimpan bukti tes. Semuanya di-ignore Git dan Docker.
- Helper API hanya lokal di .local/, memakai HTTPS IPv4 dan timeout, tidak me-retry mutasi otomatis. Respons credential disimpan lokal tanpa dicetak.
- Setelah checkpoint commit ini, baca juga .local/LAST_PROGRESS.md jika tersedia untuk hasil akhir merge/deployment yang diverifikasi sesudah CI. Catatan lokal ini tidak berisi kredensial dan tidak diunggah; berkas bukti JSON terkait tetap di-ignore.
- Screenshot daring di web/test-results/cloud-meta-report-desktop.png dan cloud-meta-review-desktop.png; sudah diperiksa visual dan di-ignore. Tidak ada trace browser yang diunggah.
- Runtime lokal: Go 1.27.3 .tools/go; Node 24; PostgreSQL portabel port 55432. Server Meta lokal lama mungkin masih port 8080; cek proses sebelum menjalankan ulang.
- opsmind_test khusus tes reset schema; jangan jalankan reset bersama server/worker pada DB test yang sama.
- Ground truth eval/ tidak dibaca runtime/model. Validasi schema/hash/referensi tidak membuktikan kebenaran semua narasi.

## Urutan lanjut yang dibatasi plan
1. P1 selesai melalui PR #2, merge, deployment, empat deteksi dan dua investigasi Meta nyata. Jangan mengulang pekerjaan ini atau mengubah golden.
2. Hover selesai melalui PR #3 dan deployment dep-dbh28pflk1mc739obg5g pada 9686cea; CI main 40727125790 lulus. Jangan mengulang pekerjaan hover.
3. Selesaikan P2 pada codex/shift-handover: tes snapshot/API/browser → commit/push/PR/CI → keputusan committer khusus PR baru → squash merge → CI main/deploy gratis/verifikasi tanpa panggilan model tambahan.
4. Demo/video ditangani pengguna sendiri. Agent tidak melanjutkan pekerjaan video atau submission tanpa instruksi berikutnya.
5. Target internal menurut plan tetap 4 November 2026 18.00 WIB; event 5 November 09.00 WIB. Jangan memperluas ke modul lain.

## P1 selesai
- Branch codex/p1-detection: ae281d1 dan perbaikan akhir 11594a2. PR #2 https://github.com/Tegar234-koplox/ops-mind-ai/pull/2 sudah squash merge menjadi 77227f80c8ac6af674abbb8d99defe9e6e66c088. Author Tegar Sang Putra, committer teknis GitHub diizinkan pengguna khusus PR #2; tanpa Co-authored-by.
- Migrasi 000004: konfirmasi absen berpasangan waktu/actor, expense_closures, approval events append-only. Data lama tidak direkayasa menjadi lengkap.
- Detektor P1 terpisah di internal/detection/p1.go; query deterministik scoped dan fixture idempotent di internal/store/p1.go dan internal/seed/p1.go. Delapan tool tetap sama.
- ENABLE_P1_RULES=false default; SEED_P1_DEMO=true memerlukan aktivasi P1. COST_VARIANCE_BPS=2000. Fixture lini B/C mempertahankan golden A.
- Kode React/TypeScript diformat dengan Prettier 3.9.10 (dev dependency terkunci), gofmt untuk Go. Pemeriksaan format ditambahkan ke CI. Tidak mengganti TanStack Query atau stack hosting.
- Tes integrasi P1, vet, suite Go lengkap, build dan format frontend lulus lokal. Browser P1 lulus (27,2 detik); regresi golden/follow-up/mobile lulus. Tes agent dan integrasi koreksi sumber lulus. CI SHA akhir PR 11594a2 lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40723750466 . CI main hasil merge lulus: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40724233043 (verify dan Container startup).
- Deskripsi delapan tool menjelaskan bukti P1. Jika koreksi sumber menghapus rangkaian ATT-001, tool menampilkan nol shift terkonfirmasi dan needs_verification, bukan mempertahankan angka pemicu lama.
- Perbaikan integrasi yang ditemukan browser: tombol deteksi kini memakai scope backend (tidak hardcode golden A); mutasi investigasi memperbarui cache latest_run secara langsung agar laporan lama tidak muncul saat follow-up berjalan.
- Screenshot/trace diarahkan ke direktori ignore; trace dinonaktifkan agar tidak diunggah oleh artifact CI.
- PostgreSQL portabel aktif port 55432 lewat scripts/portable-postgres.mjs. Jangan reset opsmind_test saat browser/server test aktif.
- Render dep-dbh20ig8v5cc73fqmk80 berstatus live pada 77227f8. Service tetap Free, flag ENABLE_P1_RULES/SEED_P1_DEMO=true. Migrasi 000004 dan fixture idempotent terpasang tanpa reset database.
- Delapan aturan aktif dan delapan kasus. ATT-001: EMP-092, satu pekerja, tiga shift. OT-002: satu catatan, 75 menit. COST-001: Rp528.999, +2999 bps karena rupiah integer. COST-002: Rp305.841, -2500 bps. Empat baseline masing-masing biaya.
- Meta nyata ATT-001 run 7c0e144ab0363c1934f3d28e1d966fb7 succeeded, dua putaran/tujuh tool/empat domain, US$0,019606. COST-002 run d47f049bf670286ed669d00cff8f7d83 succeeded, tiga putaran/enam tool/empat domain, US$0,023793. Laporan tidak menyimpulkan pelanggaran atau efisiensi pasti; rekomendasi menunggu manusia.
- Browser daring lulus: delapan aturan/kasus, laporan Meta, dialog bukti, golden +38%, tanpa pageerror. Screenshot .local/p1-cloud-report.png diperiksa visual. Run golden awal tetap succeeded dengan tujuh evidence; investigasi tambahan pengguna tetap dipertahankan.
- Saldo demo daring teramati US$0,080398 (termasuk investigasi tambahan pengguna); dua tes P1 menambah US$0,043399. Catatan lokal sebelumnya US$0,059159; gabungan teramati US$0,139557, tetap satu anggaran aplikasi US$20.

## Hover selesai
- Reproduksi browser daring membuktikan nav samping, tab, dan tautan bukti tidak memiliki perubahan hover. Tombol utama, kasus nonaktif dan kartu bukti sudah merespons; kasus terpilih tidak merespons. Bukti lokal .local/hover-before.json/png (ignore).
- PR #3 https://github.com/Tegar234-koplox/ops-mind-ai/pull/3 sudah squash merge ke 9686cea, CI lulus dan Render Free live. Browser daring membuktikan hover pada tujuh target; laporan golden dan P1 tetap utuh, saldo demo US$0,080398 tidak berubah. Kartu statistik tetap informasional.
- Tes mencakup hover normal/terpilih, dialog bukti dari tautan, aria-current, fokus keyboard, reduced motion, dan navigasi tab pada viewport ponsel. Nav samping memang disembunyikan pada layout ponsel; tidak menambah menu baru.
- Contributor repo setelah merge P1 tetap hanya Tegar234-koplox (diverifikasi API). Pengguna mengizinkan committer teknis GitHub khusus PR #3 melalui jawaban terpisah setelah PR siap ditinjau; izin tidak mencakup PR berikutnya. Merge hanya setelah seluruh checks relevan lulus.

## P2 serah terima sedang dikerjakan
- Branch codex/shift-handover dari main 9686cea. Scope dipilih eksplisit pengguna setelah ditanyakan arti P2; bukan seluruh daftar Phase 2 brief.
- Migrasi 000005 menambah shift_handovers immutable, hash dan indeks riwayat. Store reader internal mendukung transaksi repeatable-read/read-only untuk fakta serta status review yang konsisten. Pembentukan laporan Go tidak mengubah fakta/model/aturan P0/P1.
- API autentikasi: GET /shifts, GET/POST /handovers, GET /handovers/{id}. POST memakai Origin dan Idempotency-Key; shift belum selesai 409, scope fasilitas lain 404. Snapshot menyimpan empat domain, gangguan mesin, kasus, latest run, rekomendasi/review, source refs dan hash evidence.
- UI terpisah di web/src/handover, akses desktop/ponsel, filter lini/tanggal/shift selesai, riwayat snapshot, telusuri kasus dan dialog cetak/PDF browser. Tidak ada pengiriman atau penerimaan shift otomatis.
- Data_as_of adalah cutoff operasional fixture; keputusan review dibaca saat created_at laporan. Pending/cuti/closure belum lengkap tidak boleh disamakan dengan pelanggaran atau total final.
- Unit builder, suite Go/PostgreSQL lengkap dan vet lulus lokal. Browser keseluruhan: empat lulus, satu tes Meta opsional di-skip. Screenshot desktop/ponsel/print diperiksa; label produksi menggunakan bahasa Indonesia. Tes terarah terakhir lulus, termasuk perpindahan shift saat request berlangsung dan tombol yang tetap menunggu sampai penyusunan selesai. Database portabel sempat berhenti saat pergantian sesi dan dinyalakan lagi tanpa reset.
- .gitattributes mengunci LF untuk konsistensi formatter Windows/Linux; tidak mengubah konfigurasi Git global. P2 belum commit/push/PR/deploy pada checkpoint ini. Hasil lanjutan dicatat di .local/LAST_PROGRESS.md.
