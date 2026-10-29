# OpsMind AI

Platform investigasi operasional manufaktur: deteksi deterministik Go → agent Meta memilih bukti lintas departemen → laporan → review manusia. Seluruh dataset sintetis. Empat aturan P0 dan empat P1 opsional, delapan tool baca, PostgreSQL persisten, React/TypeScript, login, SSE dan audit.

## Demo daring

[OpsMind AI — demo terlindungi login](https://opsmind-ai-l0px.onrender.com). Kredensial diberikan pemilik secara privat. Seluruh catatan manufaktur sintetis. Layanan Free dapat mengalami cold start.

Verifikasi 28 Oktober 2026: login HTTPS, empat aturan P0, golden Meta nyata (2 putaran, 7 tool, 4 domain), bukti, SSE reconnect, serta review manusia berhasil. Biaya investigasi daring US$0,018857. Video dan submission belum selesai.

## Jalankan lokal dengan Docker

Prasyarat: Docker Engine/Compose aktif. Untuk pengembangan tanpa container: Go 1.27.3, Node 24, PostgreSQL 18.

1. Salin `.env.example` menjadi `.env`.
2. Isi `OPERATOR_PASSWORD` minimal 12 karakter dan `SESSION_SECRET` minimal 32 karakter acak. Jangan pakai contoh password pada deployment. Alternatif: `OPERATOR_PASSWORD_HASH` bcrypt dari `OPERATOR_PASSWORD=... go run ./cmd/opsmind hash-password`.
3. Isi `MODEL_API_KEY` untuk Meta dan pertahankan `MODEL_MODE=meta`. Untuk pengujian, pilih `MODEL_MODE=mock` secara eksplisit. Tidak ada fallback diam-diam.
4. Jalankan `docker compose up --build`.
5. Buka http://localhost:8080 dan login; tekan **Jalankan deteksi**. OT-001 otomatis masuk investigasi.

Compose memakai database lokal port 55432, frontend/API port 8080, serta seed bila kosong. Jangan jalankan database portabel dan Compose pada port yang sama. `APP_ORIGIN` harus cocok dengan origin browser; lokal `COOKIE_SECURE=false`, HTTPS `true`.

## Tanpa Docker di Windows

Go portabel pada `.tools/go` dapat dipakai jika disiapkan pada sesi pengembangan. Paket PostgreSQL portabel hanya alat pengujian lokal, bukan dependensi produksi:

```powershell
npm.cmd install --prefix .tools/pg embedded-postgres@18.4.0-beta.17
node scripts/portable-postgres.mjs
```

Pada terminal lain, set `DATABASE_URL=postgres://opsmind:local-only-database@127.0.0.1:55432/opsmind?sslmode=disable` di `.env`, lalu:

```powershell
Set-Location web
npm.cmd ci
npm.cmd run build
Set-Location ..
.\\scripts\\run.ps1 -Mock
```

`-Mock` memilih simulasi secara eksplisit untuk proses itu; berkas `.env` tetap dapat memakai default Meta. Jika konfigurasi lokal dibuat otomatis, password acak tersedia hanya di `.env`. Jangan mengunggah berkas itu. Untuk Meta nyata jalankan tanpa `-Mock` setelah mengisi key. CLI: `serve`, `migrate`, `seed`, `detect`, `hash-password`. CLI membaca environment; script PowerShell memuat `.env` tanpa mengevaluasi isinya sebagai kode.

## Golden scenario

Assembly Line A, S1, 26 Oktober 2026: biaya lembur Rp1.380.000 terhadap baseline Rp1.000.000 (+38%); output 830/1000 (-17%); 38 penugasan =31 hadir terverifikasi +3 cuti disetujui +4 pending; downtime union 110 menit. Seed 20261027, 150 pekerja, 3 lini, 8 mesin, 60 tanggal kalender. Jam evaluasi fixture: 27 Oktober 2026 06.15 WIB; ini bukan data live hari ini. Ground truth berada di `eval/` dan tidak dibaca runtime/model.

## Empat aturan P1

P1 ditambahkan atas instruksi pemilik setelah P0 stabil. Aktifkan `ENABLE_P1_RULES=true`; default tetap false. Untuk data contoh tambahan, jalankan `seed-p1` atau set `SEED_P1_DEMO=true` secara eksplisit. Fixture hanya memperluas dataset sintetis yang dikenali, tanpa mereset laporan golden. ATT-001 memerlukan konfirmasi absen, OT-002 membaca histori persetujuan, COST-001/COST-002 memerlukan baseline aktivitas dan closure posting. Threshold biaya: `COST_VARIANCE_BPS=2000`.

Gunakan `gofmt` untuk Go dan `npm run format`/`npm run format:check` di web untuk format TypeScript. Aturan P1 dipisahkan dari alur deteksi utama; migrasi, query, seed dan UI tetap pada monolit yang sama.

## Laporan serah terima shift (P2)

Buka **Serah terima shift**, pilih lini/tanggal/shift selesai lalu **Susun laporan baru**. Laporan berisi fakta empat domain dari Go, catatan yang perlu diverifikasi, gangguan mesin, kasus dan hasil investigasi/review yang sudah tersimpan. Setiap versi immutable; perubahan review tidak mengubah laporan lama. Gunakan pilihan versi untuk membuka arsip dan **Cetak / Simpan PDF** untuk dialog cetak browser. Pada ponsel, gunakan tombol pemilih ruang kerja di atas halaman.

Penyusunan tidak menambah panggilan Meta atau mengubah data operasional. Waktu evaluasi operasional tetap mengikuti fixture sintetis; status review diambil saat laporan dibuat. Snapshot bukan tanda penerimaan shift atau persetujuan tindakan. Referensi sumber dan hash membantu penelusuran, bukan jaminan bahwa semua narasi benar. Fitur memakai migrasi 000005; jangan menjalankan down migration pada database yang laporan-laporannya perlu dipertahankan.

## Pengujian aplikasi

```powershell
. .\\scripts\\load-env.ps1
$env:TEST_DATABASE_URL='postgres://opsmind:local-only-database@127.0.0.1:55432/opsmind_test?sslmode=disable'
& $GoCommand test -count=1 ./...
& $GoCommand vet ./...
& $GoCommand build -o bin/opsmind.exe ./cmd/opsmind
Set-Location web
npm.cmd ci
npm.cmd run build
npx.cmd playwright install chromium
npm.cmd run test:e2e
```

Tes integrasi **mereset schema hanya pada database TEST_DATABASE_URL yang namanya berakhiran _test**. Jangan menunjuk database bisnis. Tanpa variabel tersebut, integrasi di-skip dan harus dilaporkan sebagai skip. Linux: `go test -race ./...` dengan PostgreSQL 18; CI menyiapkan keduanya dan memakai mock, tanpa key Meta. Browser test memakai operator dari environment atau `.env`; jangan mengaktifkan mock di aplikasi publik tanpa label.

## Deployment gratis

Gunakan `render.yaml`: satu Render Free Docker web service, database Neon Free PostgreSQL 18 dengan TLS dan koneksi langsung (session lock worker memerlukan sesi PostgreSQL tetap). Gunakan `sslmode=verify-full`. Blueprint memakai region Singapore, branch main, dan deploy setelah CI lulus. Isi DATABASE_URL, APP_ORIGIN (URL deployment persis), OPERATOR_PASSWORD_HASH, MODEL_API_KEY melalui secret environment. SESSION_SECRET dihasilkan Render. BUDGET_PHASE=demo, COOKIE_SECURE=true, SEED_IF_EMPTY=true. Jangan membuat layanan berbayar atau memasang payment method tanpa keputusan pemilik.

Render Free dapat tidur setelah 15 menit tidak ada inbound traffic; cold start sekitar satu menit. Worker melanjutkan checkpoint ketika proses hidup. Semua state di PostgreSQL; filesystem container ephemeral. Ini demo dengan pendekatan produksi, bukan layanan dengan SLA. Batas aplikasi US$20 (15 pengembangan +5 demo), maksimum US$0,25/run; belum tentu seluruh kredit US$150 tersedia selamanya. Harga input/output Meta diverifikasi pada perencanaan dan harus dicek ulang sebelum rilis.

## Batas keamanan dan produk

Model tidak memiliki SQL bebas atau tool tulis. Semua tool scoped ke shift/kasus. Fakta numerik dari Go, hipotesis dari model. Hash/referensi/schema valid tidak menjamin kebenaran narasi atau sebab-akibat. Empat pending attendance tidak berarti misconduct, dan biaya abnormal tidak membuktikan fraud. Persetujuan hanya mencatat keputusan; tindakan bisnis tidak dieksekusi.

Satu akun operator environment, cookie signed 8 jam. Logout menghapus cookie browser; MVP belum memiliki pencabutan semua sesi lintas perangkat. Idempotensi memutar ulang byte respons; operasi yang crash sebelum respons tersimpan tetap pending untuk mencegah pengulangan ambigu. Periksa state kasus sebelum mengirim operasi baru. Review laporan lama ditolak. Audit append-only.

Belum termasuk ERP, predictive maintenance, public signup, multitenancy, maupun sistem produksi perusahaan.

## Dokumen dan sumber

Sepuluh dokumen perencanaan ada di `docs/`; ERD di DATABASE_DESIGN dan diagram komponen di ARCHITECTURE. HACKATHON_DEMO_PLAN berisi naskah 180 detik dan checklist submission. Tenggat internal 4 November 2026 18.00 WIB; event 5 November 09.00 WIB. Public repo, rekaman, deployment, panggilan Meta nyata dan submission hanya dinyatakan selesai setelah diverifikasi masing-masing.

- https://dev.meta.ai/events/global-hackathon
- https://dev.meta.ai/docs/tool-calling
- https://dev.meta.ai/docs/structured-output
- https://dev.meta.ai/docs/pricing-rate-limits
- https://render.com/docs/free
- https://neon.com/pricing

Karya aplikasi disusun selama periode hackathon. Library pihak ketiga mempertahankan lisensinya; periksa syarat event final pada portal peserta.

## Titik lanjut

Baca [catatan handoff](docs/HANDOFF.md) sebelum melanjutkan pekerjaan. Status pengujian, kendala publikasi, dan batas cakupan dicatat di sana.
