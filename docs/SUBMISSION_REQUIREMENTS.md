# Persyaratan submission

Checkpoint: 29 Oktober 2026. Catatan ini memisahkan informasi peserta dari hasil pemeriksaan halaman publik. Ini bukan pernyataan bahwa semua ketentuan acara telah dipenuhi atau submission telah terkirim.

## Sumber dan status verifikasi

| Ketentuan | Dasar yang tersedia | Status |
|---|---|---|
| Kategori Indie Developer | Halaman publik Meta dan Cerebral Valley; konfirmasi peserta saat onboarding | Kategori terverifikasi; pilihan kategori final pada formulir masih perlu diperiksa pemilik |
| Batas 4 November 2026 21.00 ET | Halaman publik Meta dan penyelenggara | Terverifikasi; setara 5 November 2026 09.00 WIB |
| Pekerjaan baru selama 26 Oktober–4 November 2026 | Ringkasan aturan/onboarding yang disampaikan peserta | Menjadi persyaratan kerja; pasal lengkap Official Rules belum berhasil dibaca secara independen |
| Repository GitHub publik dan dapat diakses juri | Ringkasan peserta | Repo publik sudah tersedia; akses juri perlu dicek kembali menjelang pengiriman |
| Video demonstrasi 3 menit | Ringkasan peserta | Wajib disiapkan pemilik; bahasa, format, hosting dan ketentuan durasi final perlu dicocokkan di portal |
| Submission melalui Cerebral Valley | Ringkasan peserta dan halaman acara penyelenggara | Pengiriman final belum dilakukan oleh agent |
| Pemakaian teknologi AI Meta secara substantif | Halaman publik menyebut keluarga model yang memenuhi syarat; ringkasan peserta | Implementasi memakai Muse Spark melalui Meta Model API untuk tool calling dan sintesis; bukti teknis di RELEASE_READINESS.md |
| Hak penggunaan kode, data dan aset; kepatuhan hukum/kebijakan | Ringkasan peserta | Data manufaktur sintetis; lisensi proyek dan atribusi dependensi tersedia. Pemeriksaan ketentuan final tetap diperlukan |
| Penggunaan OpenAI Codex sebagai coding assistant | Peserta tidak menemukan larangan eksplisit dalam aturan publik | Bukan persetujuan eksplisit panitia. Ketentuan lengkap penggunaan assistant/disclosure belum terverifikasi |
| Semua fitur yang diklaim harus baru, berfungsi dan dapat didemonstrasikan | Ringkasan peserta | Klaim dibatasi hasil yang benar-benar diuji; jangan menyebut fitur tertunda sebagai selesai |

Sumber primer yang diperiksa:

- [Halaman acara Meta](https://dev.meta.ai/events/global-hackathon): kategori, keluarga teknologi Meta dan tenggat.
- [Halaman penyelenggara Cerebral Valley](https://cerebralvalley.ai/e/meta-ai-developer-hackathon): acara dan tenggat.
- [Event Waiver](https://cerebralvalley.ai/e/meta-ai-developer-hackathon/waiver): halaman belum memberikan teks yang dapat dibaca pada pemeriksaan ini. Isi lengkap belum terverifikasi.

Halaman publik yang terbaca belum menyediakan seluruh pasal Official Rules/Event Waiver. Ringkasan onboarding di atas tidak menggantikan ketentuan final pada portal peserta. Target internal tetap 4 November 2026 pukul 18.00 WIB.

## Pernyataan proyek yang dapat didukung

OpsMind AI adalah agent investigasi operasional manufaktur dengan backend Go, frontend React/TypeScript dan PostgreSQL. Go menghitung anomali serta fakta numerik. Muse Spark melalui Meta Model API memilih tool baca, mengumpulkan bukti lintas departemen dan menyusun hipotesis/rekomendasi. Operator meninjau rekomendasi; persetujuan tidak menjalankan perubahan data operasional.

Codex digunakan sebagai alat bantu pengembangan kode. Engine investigasi aplikasi menggunakan Meta, bukan Codex. Seluruh data manufaktur sintetis. Ketepatan referensi dan JSON tidak membuktikan seluruh narasi benar; korelasi tidak dinyatakan sebagai sebab yang pasti. Laporan serah terima menggunakan fakta Go dan hasil investigasi/review tersimpan tanpa panggilan model baru.

Repo menyimpan riwayat implementasi dan pengujian; riwayat tersebut membantu penelusuran pekerjaan, tetapi bukan sertifikasi kelayakan peserta. Library pihak ketiga diidentifikasi dalam [atribusi](../THIRD_PARTY_NOTICES.md). Jangan mengklaim library tersebut sebagai karya baru peserta.

## Pemeriksaan pemilik sebelum pengiriman

- [ ] Baca dan simpan rujukan Official Rules serta Event Waiver final dari portal; cocokkan periode pekerjaan, kategori/kelayakan, kepemilikan, penggunaan assistant dan kewajiban disclosure.
- [ ] Pastikan seluruh klaim fitur cocok dengan hasil pengujian; periksa tanggal pekerjaan dan izin komponen/aset yang digunakan.
- [ ] Buka repository publik melalui sesi tanpa login. Jangan mengunggah .env, password, API key, DSN atau artefak runtime privat.
- [ ] Rekam dan periksa video tiga menit; tunjukkan investigasi Meta nyata, bukti, ketidakpastian dan review manusia. Bedakan laporan tersimpan, rekaman cadangan dan mock.
- [ ] Uji URL aplikasi dan berikan kredensial juri melalui kanal submission privat. Jelaskan cold start layanan gratis.
- [ ] Lengkapi formulir Cerebral Valley, periksa semua tautan/lampiran dan simpan tanda terima sebelum tenggat.

Demo, video, keputusan terhadap ketentuan acara dan submission ditangani pemilik. Agent tidak menyatakan pengiriman selesai tanpa bukti tanda terima.
