# Handoff OpsMind AI

Diperbarui 28 Oktober 2026 (WIB). Status: pekerjaan dilanjutkan atas instruksi pengguna.

## Cakupan dan keputusan yang tetap dikunci
- Monolit modular Go/PostgreSQL/React/Meta Model API; empat aturan P0: OT-001, PROD-001, ATT-002, MAINT-001.
- P1 dan ERP, microservices, multi-fasilitas, prediksi, serta registrasi publik tetap ditunda.
- Data sintetis; 4 attendance pending bukan bukti mangkir. Review tidak menulis data bisnis.
- Anggaran tunai US$0. Kredit aplikasi US$20 (15 development/5 demo), US$0,25/run. Jangan membeli layanan atau menaikkan paket.
- Author dan committer hanya Tegar Sang Putra / Tegar234-koplox, tanpa GPT/bot atau trailer Co-authored-by.
- Alur Git berikutnya: branch codex/ → commit → push branch → PR → squash and merge saat siap dan pemeriksaan lulus. Jangan push perubahan berikutnya langsung ke main.

## Checkpoint yang sudah terverifikasi
- Commit aplikasi awal 3d6ef8aca859df8b6d37b570335caba43c920d13 sudah berhasil dipublikasikan ke main:
  https://github.com/Tegar234-koplox/ops-mind-ai
- SHA lokal dan remote identik; metadata author dan committer akun pengguna.
- CI jarak jauh pertama LULUS: https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40602651746
- CI tersebut menjalankan go vet, go test -race dengan PostgreSQL 18, build frontend, dan Playwright.
- Verifikasi lokal sebelumnya juga lulus: golden Meta nyata 2 putaran/7 tool/7 evidence/2 hipotesis/3 rekomendasi; run 4942abc42683888ec04fb6de8e6f47da.
- Biaya run berhasil US$0,019756. Total lokal terpakai/direservasi pada checkpoint lama US$0,059159, termasuk reservasi konservatif percobaan jaringan.
- Angka golden: +38% lembur; 830/1000 output (-17%); 31 hadir +3 cuti +4 pending =38; downtime union 110 menit.
- Satu rekomendasi nyata disetujui sebagai uji data sintetis, dua pending; tidak ada tindakan bisnis nyata.
- Perbaikan terdahulu: idempotensi response bytea, logout cache observer, batas completion Meta 4096 dengan reasoning low dan satu kesempatan repair. Reasoning tidak dikirim ke UI.

## Pekerjaan sesi ini
- Branch codex/deployment-readiness disiapkan untuk pemeriksaan startup container dan pembaruan petunjuk Git.
- Docker lokal masih belum menyediakan daemon walaupun Desktop sudah diminta berjalan. Validasi container dipindahkan ke job CI tersendiri pada runner Linux.
- Job tersebut membangun Dockerfile, menjalankan Compose/PostgreSQL, memeriksa health/frontend/login/seed 150 pekerja, lalu membersihkan container pengujian.
- Status job baru dan PR harus dikonfirmasi setelah push; jangan mengklaim lulus sebelum hasil aktual tersedia.

## Akses deployment
- Pengguna telah menyelesaikan otorisasi Neon dan Render.
- Konektor baru sempat menampilkan daftar alat tetapi pemanggilan gagal Unknown tool, kemudian alatnya tidak tersedia lagi pada sesi. Ini kendala integrasi, bukan bukti pengguna belum memberi izin.
- Pengguna memilih jalur API key lokal. Template .env.deploy sudah dibuat dan di-ignore Git/Docker; isi NEON_API_KEY dan RENDER_API_KEY hanya lokal. Jangan cetak nilai atau commit.
- Belum ada project Neon, service Render, atau biaya hosting yang dibuat oleh agent.
- Sebelum provisioning: baca organisasi/proyek Neon dan workspace/service Render; pilih/reuse resource yang sesuai, konfirmasi plan free, dan hindari duplikasi. Jangan downgrade atau mengubah resource lain.
- Neon baru harus memakai PostgreSQL 18, compute kecil, suspend saat idle. Render harus plan free, Docker, branch main, HTTPS cookie/origin, dan BUDGET_PHASE=demo.
- Pindahkan MODEL_API_KEY hanya ke secret environment backend Render; tidak ke repo, frontend, logs atau PR.

## Solusi transport Git yang berhasil
Git langsung sebelumnya gagal karena kombinasi backend TLS openssl tidak didukung executable helper dan jalur IPv6/NAT64 yang reset. Perintah per operasi berikut berhasil tanpa mengubah Git global atau pengaturan jaringan sistem:
git -c http.sslBackend=schannel -c http.curloptResolve=github.com:443:20.205.243.166 -c credential.helper= -c 'credential.helper=!gh auth git-credential' ...
- Alamat IPv4 itu diverifikasi dari DNS/curl pada sesi ini; jika berubah, resolve ulang dari github.com dengan TLS tetap diverifikasi.
- GitHub CLI sudah login akun Tegar234-koplox, scopes repo dan workflow. Jangan tampilkan token.
- Push aplikasi awal ke main terjadi sebelum instruksi workflow PR baru. Semua perubahan selanjutnya harus melalui branch dan PR.
- Metadata dan CI dapat dibaca lewat konektor GitHub; izin tulis konektor sebelumnya 403, sehingga gunakan transport CLI akun pengguna.

## Catatan runtime dan dataset
- Go portabel 1.27.3 di .tools/go; Node 24; PostgreSQL portabel melalui scripts/portable-postgres.mjs pada port 55432.
- .env berisi operator password, SESSION_SECRET dan MODEL_API_KEY pengguna. Tidak boleh dicetak/commit/upload.
- Database opsmind untuk demo; opsmind_test khusus tes yang mereset schema. Jangan menjalankan integration reset saat server lain memakai database test.
- Database demo lokal lama dibuat sebelum generator menyamakan tarif lembur menjadi Rp500/menit; angka golden tetap sama. Deployment harus memakai database baru dari generator terakhir untuk konsistensi semua catatan.
- Server lokal Meta sebelumnya port 8080 dari bin/opsmind-next.exe; cek proses/health sebelum mencoba menjalankan ulang.
- Screenshot sebelumnya ada di web/test-results (di-ignore), dapat terhapus saat test berikutnya.
- Ground truth di eval/ tidak dibaca runtime/model. Schema/hash/referensi valid bukan bukti kausalitas.

## Urutan lanjut
1. Commit/push perubahan readiness ke branch, buat PR, periksa job container dan pemeriksaan standar yang relevan.
2. Setelah key lokal tersedia, baca akun/plan; provision hanya Neon Free dan Render Free sesuai arsitektur.
3. Uji deployment dengan dataset baru, Meta nyata, SSE/reconnect, dan review; pastikan tidak ada secret di sumber publik.
4. Perbarui status handoff dengan URL, hasil aktual, dan kendala yang masih ada. Squash merge PR ketika siap sesuai workflow pengguna.
5. Finalisasi video 180 detik, akses penilai dan rules final. Tenggat internal 4 November 18.00 WIB, event 5 November 09.00 WIB.
