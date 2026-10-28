# OpsMind AI — Product Requirements

Status: arsitektur disetujui melalui instruksi implementasi pengguna, 27 Oktober 2026.

## Masalah dan pengguna
Manajer operasi melihat kenaikan lembur, produksi yang menurun, dan data pekerja yang tidak lengkap di sistem berbeda. OpsMind mengumpulkan bukti tersebut menjadi investigasi yang dapat ditelusuri. Pengguna utama adalah manajer operasi; supervisor lini, HR, keuangan, dan maintenance menjadi peninjau bukti. Seluruh data demo sintetis.

## Alur dan keberhasilan
Deteksi → investigasi mandiri → bukti lintas departemen → hipotesis → rekomendasi → keputusan manusia. Empat tampilan: Inbox Kasus, Investigasi Langsung, Laporan Bukti, Review Manusia. Angka fakta dihitung Go; Meta memilih tool dan menyusun hipotesis. Persetujuan hanya merekam keputusan, tidak menjalankan tindakan operasional.

Keberhasilan: golden scenario selesai dengan bukti dari workforce, biaya, produksi, maintenance; 31 hadir + 3 cuti + 4 belum terverifikasi = 38 terjadwal; biaya lembur naik tepat 38%; produksi kurang tepat 17%; downtime lini 110 menit tanpa penghitungan ganda. Semua rekomendasi memiliki bukti dan dapat diterima atau ditolak dengan audit. Investigasi yang datanya tidak cukup menyatakan kekurangan itu.

## Prioritas
P0: OT-001, PROD-001, ATT-002, MAINT-001; delapan tool baca; laporan tervalidasi; login; antrean persisten; review dan audit. P1: ATT-001, OT-002, COST-001, COST-002, diimplementasikan setelah P0 stabil atas instruksi pengguna 28 Oktober 2026, aktivasi eksplisit melalui ENABLE_P1_RULES. Ditunda: ERP, procurement, warehouse, quality control, predictive maintenance, multi-fasilitas, registrasi publik, penulisan otomatis ke sistem bisnis.

## Batas
Satu pembangun, 3–5 jam/hari, sekitar 36 jam. Anggaran tunai US$0 dan kredit Meta US$150 menurut pengguna; anggaran aplikasi US$20 (15 pengembangan, 5 demo), maksimum US$0,25 per investigasi. Kredit bukan alasan menambah cakupan. Hosting gratis tidak memberikan jaminan ketersediaan produksi.

Target end-to-end: 31 Oktober. Pengiriman internal: 4 November 18.00 WIB. Tenggat event: 5 November 09.00 WIB. Repositori publik dan video 180 detik adalah persyaratan proyek; format, bahasa submission, lisensi dan persyaratan final harus diperiksa di portal peserta. Sumber: https://dev.meta.ai/events/global-hackathon.
