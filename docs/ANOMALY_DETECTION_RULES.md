# Aturan Deteksi

Semua deteksi Go; waktu evaluasi eksplisit, bukan jam model. Kasus memuat rule/version, entity, department, shift, expected, observed, severity, data_quality, source refs dan explanation. Aturan unik per entity/periode/versi agar rerun tidak menggandakan kasus. Threshold dari environment, dicatat bersama kasus. Perbandingan strict lebih besar; tepat threshold tidak memicu.

| Aturan | Prioritas | Definisi |
|---|---|---|
| OT-001 | P0 | Variance biaya lembur terhadap median 4 shift lengkap sebanding >2000 bps. Baseline dinormalisasi target terencana, bukan output aktual. Baseline nol/kurang → insufficient evidence, tanpa pembagian. |
| PROD-001 | P0 | (planned-actual)/planned >1000 bps setelah shift selesai. Target nol → tidak dihitung; catatan belum lengkap → perlu verifikasi. |
| ATT-002 | P0 | Punch wajib hilang atau belum terverifikasi setelah end+15 menit, tidak sedang cuti disetujui/holiday. Hasil needs_verification, bukan confirmed absence. |
| MAINT-001 | P0 | Union interval downtime yang di-clamp ke batas shift pada lini >60 menit. Overlap lintas mesin tidak dihitung dua kali sebagai downtime lini. |
| ATT-001 | P1, belum aktif | Tiga shift terjadwal berturut-turut tanpa kehadiran, cuti/holiday sudah dicek, sumber dipastikan lengkap. Tidak menyimpulkan misconduct. |
| OT-002 | P1, belum aktif | Lembur aktual positif tanpa approval valid; pembatalan approval mempertahankan audit. |
| COST-001 | P1, belum aktif | Pengeluaran melebihi baseline aktivitas sebanding >20%; biaya per unit dan perubahan kondisi dianalisis terpisah. |
| COST-002 | P1, belum aktif | Pengeluaran di bawah baseline aktivitas >20%, dengan konfirmasi posting lengkap; missing invoices adalah gap. |

OT-001 high pada >=3500 bps, selain itu medium. PROD-001 high pada >=2000 bps, selain itu medium. MAINT-001 high pada >=120 menit, selain itu medium. ATT-002 low dan needs_verification. insufficient_evidence tidak diberi label anomali terkonfirmasi.

Golden: OT 1.000.000 →1.380.000 =3800 bps/high; produksi 1.000→830 =1700 bps/medium; downtime 110 menit/medium; empat pending attendance/low. Referensi attendance cuti yang sah tidak menjadi kasus punch hilang.
