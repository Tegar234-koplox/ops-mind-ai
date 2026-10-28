# Aturan Deteksi

Semua deteksi Go; waktu evaluasi eksplisit, bukan jam model. Kasus memuat rule/version, entity, department, shift, expected, observed, severity, data_quality, source refs dan explanation. Aturan unik per entity/periode/versi agar rerun tidak menggandakan kasus. Threshold dari environment, dicatat bersama kasus. Perbandingan strict lebih besar; tepat threshold tidak memicu.

| Aturan | Prioritas | Definisi |
|---|---|---|
| OT-001 | P0 | Variance biaya lembur terhadap median 4 shift lengkap sebanding >2000 bps. Baseline dinormalisasi target terencana, bukan output aktual. Baseline nol/kurang → insufficient evidence, tanpa pembagian. |
| PROD-001 | P0 | (planned-actual)/planned >1000 bps setelah shift selesai. Target nol → tidak dihitung; catatan belum lengkap → perlu verifikasi. |
| ATT-002 | P0 | Punch wajib hilang atau belum terverifikasi setelah end+15 menit, tidak sedang cuti disetujui/holiday. Hasil needs_verification, bukan confirmed absence. |
| MAINT-001 | P0 | Union interval downtime yang di-clamp ke batas shift pada lini >60 menit. Overlap lintas mesin tidak dihitung dua kali sebagai downtime lini. |
| ATT-001 | P1, opt-in | Tiga penugasan kerja terakhir memiliki konfirmasi absen supervisor dan sumber lengkap. Hari libur dilewati; cuti pada salah satu penugasan memutus rangkaian. Punch kosong/pending tidak cukup. |
| OT-002 | P1, opt-in | Menit lembur aktual >0 tanpa persetujuan efektif. Kejadian approval terakhir sampai waktu evaluasi mengalahkan snapshot approved; riwayat approval/revocation append-only. Catatan belum lengkap tetap perlu verifikasi. |
| COST-001 | P1, opt-in | Total pengeluaran >20% di atas median empat shift lengkap sebanding, dinormalisasi target terencana. Closure posting wajib. Biaya per unit/perubahan kondisi dianalisis terpisah. |
| COST-002 | P1, opt-in | Total pengeluaran >20% di bawah baseline yang sama, setelah posting ditutup. Missing invoice, closure dini/future, baseline nol/kurang tidak menghasilkan klaim pengeluaran rendah terkonfirmasi. |

OT-001 high pada >=3500 bps, selain itu medium. PROD-001 high pada >=2000 bps, selain itu medium. MAINT-001 high pada >=120 menit, selain itu medium. ATT-002 low dan needs_verification. insufficient_evidence tidak diberi label anomali terkonfirmasi.

Golden: OT 1.000.000 →1.380.000 =3800 bps/high; produksi 1.000→830 =1700 bps/medium; downtime 110 menit/medium; empat pending attendance/low. Referensi attendance cuti yang sah tidak menjadi kasus punch hilang.

## Aktivasi dan spesifikasi P1

P1 dikerjakan atas instruksi pengguna 28 Oktober 2026. `ENABLE_P1_RULES=true` mengaktifkan keempat aturan bersama; default false mempertahankan perilaku P0. Threshold biaya `COST_VARIANCE_BPS=2000`, dihitung Go sebagai basis points integer; tepat +2000/-2000 tidak memicu. COST high pada nilai absolut >=3500 bps, selain itu medium. ATT-001 medium; OT-002 medium dan needs_verification jika terdapat sumber lembur belum lengkap.

Baseline biaya: empat shift terdahulu pada lini, kode shift, produk, durasi dan kelas hari kerja yang sama; target positif, produksi dan expense lengkap, ledger ditutup sesudah akhir shift dan sebelum/saat waktu evaluasi. Semua kategori expense dijumlahkan, tanpa memakai output aktual sebagai normalisasi. Jika baseline tidak evaluable atau ledger belum ditutup, COST tidak dikonfirmasi; tool menjelaskan insufficient_evidence. Nilai rendah belum membuktikan efisiensi.

ATT-001 diagregasi per lini/shift dengan ID pekerja sintetis dan referensi tiga penugasan terkait. Konfirmasi harus terjadi sesudah akhir shift, sebelum/saat evaluasi, memiliki actor dan source_complete=true. Approval OT memakai waktu evaluasi yang sama; jika belum ada event, snapshot approved yang sudah ada tetap menjadi sumber. Tidak ada endpoint untuk mengubah approval atau kehadiran.

Fixture P1 eksplisit: B-20261026-S1 untuk ATT-001, B-20261026-S2 untuk OT-002, B-20261026-S3 untuk COST-001, C-20261026-S3 untuk COST-002. Jalankan `seed-p1` atau `SEED_P1_DEMO=true` bersama ENABLE_P1_RULES. Seed transactional/idempotent dan hanya menerima dataset sintetis 150 pekerja yang dikenali. Golden A-20261026-S1 tetap empat kasus P0 dengan angka yang sama.
