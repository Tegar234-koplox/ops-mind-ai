# Handoff OpsMind AI

Diperbarui 29 Oktober 2026. P0, P1, hover dan P2 serah terima shift selesai serta diverifikasi daring. Tahap aktif: kesiapan rilis, bukan modul bisnis baru. Baca juga .local/LAST_PROGRESS.md jika tersedia untuk hasil commit/CI/merge/deploy yang terjadi setelah checkpoint dokumen ini.

## Keputusan tetap

- Monolit modular Go/PostgreSQL/React/Meta Model API; arsitektur sudah disetujui.
- P0: OT-001, PROD-001, ATT-002, MAINT-001 dan golden lengkap. P1 diaktifkan pengguna pada 28 Oktober: ATT-001, OT-002, COST-001, COST-002.
- P2 hanya laporan serah terima shift: snapshot immutable atas permintaan, fakta Go dan investigasi/review tersimpan, riwayat/cetak. Tidak ada panggilan model baru, scheduler, pengiriman, penerima atau perubahan data bisnis.
- Procurement, warehouse, quality control, predictive maintenance, ERP, multi-fasilitas, registrasi publik dan layanan tambahan tetap ditunda.
- Data manufaktur seluruhnya sintetis. Empat pending bukan bukti mangkir; tiga cuti disetujui dikecualikan. Korelasi bukan kausalitas; review tidak menjalankan rekomendasi.
- Tunai US$0; batas aplikasi US$20 (15 pengembangan/5 demo), maksimum US$0,25/run. Counter lokal dan daring bukan dua anggaran tambahan.
- Demo/video/submission ditangani pengguna. Kredensial privat di .local/DEMO_ACCESS.md; jangan cetak atau upload.
- Git: branch codex/ → commit atas nama pengguna → push branch → PR → squash merge setelah checks. Jangan push perubahan berikutnya ke main.
- Author/contributor tetap Tegar Sang Putra / Tegar234-koplox, tanpa GPT/bot/Co-authored-by. Pengecualian committer teknis GitHub hanya diizinkan terpisah untuk PR #1–#4; bukan izin umum.

## Rilis fungsional terverifikasi

- Repo publik: https://github.com/Tegar234-koplox/ops-mind-ai .
- PR #1 P0/deployment, #2 P1, #3 hover dan #4 handover sudah squash merge. Main hasil PR #4: b6eaf5be0148b7faa348665ffb611ed375759e53.
- CI PR #4 final ca0593e: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40904343526 . CI main b6eaf5b: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40904910986 . Keduanya lulus; Meta opsional di-skip secara sengaja di CI.
- CI mencakup Go format/vet/race/PostgreSQL 18, format/build frontend, browser mock dan startup container. Docker lokal sebelumnya belum memiliki daemon yang bisa digunakan; CI Linux memvalidasi container nyata.
- Render deployment dep-dbhmm2e7bikc73f15eh0 live pada b6eaf5b: https://opsmind-ai-l0px.onrender.com . Service tetap Free, Neon/database lama dipertahankan tanpa reset.
- Model daring meta/muse-spark-1.3, fase demo; delapan aturan aktif. Cookie HTTPS Secure/HttpOnly/SameSite Strict, API tanpa login 401 dan origin asing 403 terverifikasi.
- Deployment dipicu API setelah CI main lulus. checksPass terkonfigurasi; pemicuan otomatis CI-to-Render belum diverifikasi.
- Contributor API setelah merge hanya Tegar234-koplox. Author commit branch akun pengguna; committer teknis merge hanya sesuai pengecualian PR terkait.

## Bukti yang harus dipertahankan

Golden A-20261026-S1: Rp1.380.000 vs baseline Rp1.000.000 (+38%), output 830/1000 (-17%), 38 =31 hadir +3 cuti +4 pending, downtime union 110 menit. Seed 20261027: 150 pekerja, tiga lini, delapan mesin, 60 tanggal kalender; cutoff fixture 27 Oktober 06.15 WIB.

- Golden case 6cf2a0b1f6e4a08521ec6619cb960c2c; run fc4ed857586151f2e08aed563cbf4dbe succeeded, Meta nyata, dua putaran/tujuh tool/tujuh bukti/empat domain. Satu rekomendasi approved pada verifikasi awal, lainnya pending. Review tambahan pengguna tetap dipertahankan.
- P1 Meta nyata: ATT-001 run 7c0e144ab0363c1934f3d28e1d966fb7; COST-002 run d47f049bf670286ed669d00cff8f7d83. Keduanya succeeded dan mencakup empat domain.
- P2 snapshot golden b76fce1612bca7d0715d838ed9f1c2e1: empat kasus/319 sumber. Snapshot B-20261026-S1 72724a4dbaf141e0bd2ba6ec8f5b5b25: satu kasus ATT-001/83 sumber. Persistensi, hash dan replay idempotensi terverifikasi.
- Saldo demo sebelum/sesudah P2 tetap US$0,115335; tanpa panggilan Meta tambahan. Catatan lokal sebelumnya US$0,059159, gabungan teramati US$0,174494. Ini checkpoint historis, bukan saldo yang dijamin tetap.
- Laporan/bukti/review golden dan P1 tetap utuh setelah deployment P2. Browser daring memeriksa sintesis tersimpan, cetak, ponsel, penelusuran kasus dan nol pageerror.
- CI P2 awal menemukan race pada reset mutasi saat pindah shift. Sudah diperbaiki pada head ca0593e: status pending dipertahankan, error dibatasi shift terkait dan cache riwayat diisi dari hasil baru. Tes request tertunda dan CI final lulus.

## Kesiapan rilis yang sedang diselesaikan

Branch codex/release-readiness dari b6eaf5b. Pengujian HTTP provider tambahan lulus lokal pada opsmind_release_test: retry 429/503, 401, timeout, JSON/usage rusak, truncation/filter, redirect, batas budget/payload, tool/scope asing serta bukti kurang. Seluruhnya HTTP simulasi berlabel mock; tidak menghubungi Meta nyata.

README, diagram, [matriks bukti](RELEASE_READINESS.md), [persyaratan submission](SUBMISSION_REQUIREMENTS.md) dan atribusi dependensi diselaraskan dengan hasil nyata. Salinan lisensi disertakan dalam kemasan runtime. CI/merge/deploy perubahan tahap ini harus diverifikasi pada head PR-nya; belum diklaim selesai oleh checkpoint ini.

Ringkasan peserta/onboarding menjadi persyaratan kerja: Indie Developer, pekerjaan baru 26 Oktober–4 November, public repo, video tiga menit dan submission Cerebral Valley. Halaman publik mengonfirmasi kategori/keluarga teknologi Meta/tenggat; pasal lengkap Official Rules/Event Waiver belum terbaca. Codex tidak dianggap mendapat persetujuan eksplisit panitia. Baca SUBMISSION_REQUIREMENTS sebelum submission.

## Runtime dan keamanan kerja

- .env menyimpan key Meta, password/session lokal. .env.deploy menyimpan kunci management Neon/Render. .local/cloud-state.json mengandung DSN rahasia. Jangan cetak/commit/upload.
- .local menyimpan helper API, bukti verifikasi, screenshot dan checkpoint; .tools menyimpan runtime lokal. Keduanya di-ignore Git/Docker. Ground truth eval/ tidak dibaca runtime/model.
- PostgreSQL portabel port 55432; Go 1.27.3 pada .tools/go; Node 24. Periksa proses sebelum menyalakan ulang database. Pemeriksaan proses/socket di sandbox dapat berbeda dari konteks elevated; jangan menghapus postmaster.pid hanya karena pemeriksaan sandbox gagal.
- Integration test mereset schema hanya TEST_DATABASE_URL berakhiran _test. Jangan jalankan bersamaan server/worker pada database test yang sama. opsmind_release_test dipakai tes terarah tahap ini.
- Neon menggunakan koneksi langsung dan sslmode=verify-full, bukan transaction pooler karena session advisory lock. Render Free dapat tidur; tidak ada SLA 24/7.
- Git transport IPv4 scoped yang pernah berhasil: schannel + curloptResolve github.com:443:20.205.243.166 + helper gh auth git-credential, tanpa perubahan global. Token tidak dicetak.
- Koneksi Chromium daring pernah reset; helper lokal memakai lookup IPv4 sistem dan launch arguments scoped, tanpa melewati validasi TLS.

## Urutan lanjut

Selesaikan PR/CI kesiapan rilis dan alur merge sesuai batas identitas; bila paket runtime berubah, deploy gratis dan periksa persistensi tanpa menambah panggilan Meta. Catat hasil akhir di .local/LAST_PROGRESS.md. Jangan mengulang seluruh tes tanpa perubahan/failure baru.

Setelah itu sisa utama milik pengguna: pemeriksaan rules/waiver final, video, akses juri dan submission. Target internal 4 November 2026 18.00 WIB; batas acara 5 November 09.00 WIB. Jangan otomatis membuka P3 atau modul tertunda.
