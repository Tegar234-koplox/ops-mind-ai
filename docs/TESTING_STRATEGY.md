# Strategi Pengujian

## Unit
Aturan tepat/di bawah/di atas threshold; median baseline target-adjusted dan minimum empat; denominator nol; rupiah besar tanpa float; interval union/clamp; shift overnight; cuti mencakup shift, holiday, grace period, pending source. Validasi tool allowlist/parameter asing/scope, evidence palsu/lintas kasus, laporan JSON invalid, klasifikasi fakta/hipotesis, limits, retry dan budget.

## Integrasi PostgreSQL
Jalankan migrasi pada database khusus TEST_DATABASE_URL; tidak pernah memakai DATABASE_URL produksi untuk reset. Seed dua kali tidak menggandakan data. Constraint dan query, angka golden, worker lock/lease, checkpoint resume, pending tool, idempotency replay/conflict, budget reservasi lintas proses, rekomendasi usang, review dan audit transactional. Uji restart memutus worker setelah tool tersimpan lalu melanjutkan tanpa menggandakan eksekusi.

## Evaluasi agent
MODEL_MODE=mock deterministik memakai tool loop asli dan label simulasi. Skenario: golden berhasil, sumber tidak lengkap, tool tak dikenal, argumen invalid, model invalid JSON, referensi palsu, timeout, 429/5xx, pembatalan, budget habis. Golden wajib evidence dari 4 domain dan approved leave; empat pekerja tetap belum terverifikasi. Ground truth di eval/ hanya untuk evaluator; aplikasi/model tidak membaca file tersebut. Keberadaan fixture evaluasi tidak berarti ada runner evaluasi terpisah yang sudah dijalankan. JSON valid bukan jaminan kebenaran semantik. Meta smoke test opsional dan terpisah dari CI; tidak membocorkan key dalam output.

## Browser
Login gagal/berhasil, inbox, deteksi, progres SSE, refresh/resume, laporan dan detail sumber, approved/rejected beserta alasan, follow-up, pembatalan dan logout. Klik ganda tidak menciptakan run/review duplikat. Layar kecil tetap dapat membaca bukti; form memiliki label dan status async dapat diumumkan.

## Perintah dan penerimaan
go test ./...; go vet ./...; go test -race ./... bila compiler C tersedia; TEST_DATABASE_URL=... go test ./internal/store ./internal/httpapi ./internal/agent; di web: npm ci, npm run build, npm run test:e2e. CI menyediakan PostgreSQL 18 dan mock; tidak menggunakan Meta key. Test integrasi yang skip tanpa TEST_DATABASE_URL harus dilaporkan sebagai skip, bukan lulus.

Lulus MVP: angka fixture tepat; 4 aturan; semua tool scoped; laporan tervalidasi; review diaudit; state selamat restart; batas biaya benar; frontend terbuild; alur browser selesai. Deployment dan Meta nyata hanya disebut terverifikasi jika benar-benar dijalankan, bukan karena config tersedia.

## Pengujian P1
Tes upgrade v3 berisi data → v4 memastikan tidak ada closure/konfirmasi buatan. Fixture idempotent, pengecualian cuti/libur, konfirmasi future/pending, revocation append-only/future, threshold biaya tepat +/-20%, baseline kurang/nol, invoice belum lengkap dan scope historis diuji lewat PostgreSQL. P1 agent memakai mock eksplisit di CI. Browser p1.spec.ts memeriksa empat aturan, bukti rangkaian dan golden +38% yang tetap utuh. CI juga memeriksa format TypeScript. Meta nyata tetap terpisah.

## Pengujian P2 serah terima

Builder: uang tetap string integer, target nol tidak menghasilkan persentase, baseline kurang dan closure belum ada tetap perlu verifikasi. Hash berubah bila fakta berubah dan stabil setelah round-trip JSON. PostgreSQL/API: migrasi upgrade, empat fakta golden, shift malam, shift belum selesai, scope fasilitas, keputusan review saat snapshot, immutability, audit, auth/origin, idempotency replay/conflict, pagination dan tidak bertambahnya run/reservasi model. Browser: pilih lini/tanggal/shift, susun laporan, buka versi lama sesudah reload, telusuri kasus, cetak, dan akses viewport ponsel. CI memakai mock eksplisit; verifikasi daring menggunakan hasil Meta yang sudah ada tanpa panggilan baru.

## HTTP provider dan batas kegagalan

TestSimulatedMetaHTTPBoundary di internal/integration/agent_boundary_test.go memakai httptest dan database khusus berakhiran _test. Kasus HTTP: sukses dan pencatatan usage, 429/503 retryable, 401 non-retryable, JSON rusak, usage hilang/negatif, choice kosong, truncation, content filter, timeout dan redirect yang tidak diikuti. Budget/payload yang ditolak tidak boleh mengirim HTTP; timeout mempertahankan reservasi konservatif.

Runner diuji dengan model simulasi: maksimum dua retry sementara, satu repair respons invalid/truncated, tool tak dikenal/scope shift asing ditolak, deadline dihentikan, dan laporan dari satu domain menjadi insufficient_evidence tanpa rekomendasi. Panggilan HTTP simulasi tetap berlabel mock dalam database. Tes terarah ini lulus lokal pada 29 Oktober; tidak menghabiskan kredit Meta atau mengubah database daring. CI tetap menjalankan suite penuh pada head PR.

Bukti rilis fungsional dan batas cakupan pengujian ada di RELEASE_READINESS.md. Uji simulasi bukan pembuktian seluruh perilaku provider nyata; pengujian Meta nyata sebelumnya dilaporkan terpisah.
