# Persyaratan Manufaktur

## Waktu dan rekonsiliasi
Fasilitas sintetis FACTORY-01 memakai Asia/Bangkok (UTC+7), mata uang IDR. Shift S1 06.00–14.00, S2 14.00–22.00, S3 22.00–06.00 esok hari. Tanggal operasional adalah tanggal mulai shift. Semua timestamp database timestamptz. Interval memakai batas [mulai, selesai); durasi dan overlap dihitung Go. Karyawan tidak boleh ditugaskan ke dua shift bertumpang tindih.

Kehadiran lengkap berarti kedua punch terverifikasi dan urut. Cuti yang disetujui dan menutupi shift dikecualikan dari kewajiban punch. Catatan pending atau punch yang hilang adalah kekurangan bukti, bukan bukti mangkir. Holiday dan shift belum selesai tidak dihitung sebagai absensi. Grace period sesudah akhir shift: 15 menit. Tiga absen berturut-turut berarti tiga shift kerja terjadwal terakhir, bukan tiga tanggal kalender; ATT-001 diaktifkan opsional pada tahap P1. Tiga penugasan tersebut harus memiliki konfirmasi absen supervisor, bukan sekadar punch kosong. Cuti memutus rangkaian; hari libur tidak masuk urutan.

## Golden scenario
Assembly Line A, shift S1 pada 26 Oktober 2026. Sebanyak 38 orang ditugaskan: 31 terverifikasi hadir, tiga cuti disetujui, empat catatan pending (kedua punch belum terverifikasi). Tidak ada kesimpulan pelanggaran pekerja. Produksi target 1.000, aktual 830. Biaya lembur Rp1.380.000 dibanding median empat shift pembanding Rp1.000.000: +38%. Downtime lini gabungan 110 menit. Ketiga cuti tetap merupakan faktor ketersediaan staffing yang sah, bukan kesalahan pekerja.

## Baseline dan uang
Pembanding: empat shift sebelumnya pada lini, kode shift dan produk yang sama, rencana durasi sama, kelas hari kerja sama, serta sumber lengkap. Biaya pembanding dinormalisasi menurut rasio target produksi; median dihitung dengan rasional eksak. Minimum empat pembanding. Actual output tidak menjadi denominator baseline lembur. Uang disimpan sebagai BIGINT rupiah, JSON string, tanpa float untuk perhitungan uang. Variance memakai basis points (100 bps = 1%). Pembagian memakai big.Rat, pembulatan sekali pada batas representasi.

## Dataset
Seed 20261027, 60 tanggal operasional 28 Agustus–26 Oktober, evaluasi per 27 Oktober 2026 06.15 WIB. 150 karyawan sintetis: 120 operator, 30 staf pendukung; tiga lini dan delapan mesin (3/3/2). S1 lini A 38 operator, S2/S3 10; lini B/C masing-masing 8 per shift. Hari Minggu merupakan libur fasilitas sintetis. Kapasitas tidak disamaratakan ke 38 orang pada semua shift. Catatan missing, cuti, overlap downtime dan hari libur menjadi kasus uji. Label ground truth berada di direktori evaluasi dan tidak pernah diberikan ke model.

## Batas interpretasi
Kenaikan biaya bukan fraud; absensi belum jelas bukan misconduct; tidak menyimpulkan kondisi kesehatan atau kejujuran. Downtime dan staffing yang bersamaan dengan penurunan produksi adalah hipotesis kontribusi, bukan bukti kausal. Maintenance MVP hanya downtime dan konteks mesin; work order serta jadwal preventive maintenance ditunda.

## Sumber tambahan P1
Konfirmasi absen menyimpan waktu dan actor; keduanya harus ada dan catatan sumber lengkap. Pembatalan persetujuan lembur disimpan sebagai kejadian baru tanpa menghapus approval sebelumnya. Closure ledger expense tidak otomatis dibuat oleh migrasi; posting nol harus dikonfirmasi dengan closure yang sah. Fixture sintetis P1 dijalankan eksplisit dan tidak mereset golden atau investigasi tersimpan.
