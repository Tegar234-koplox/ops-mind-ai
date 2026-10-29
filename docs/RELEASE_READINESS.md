# Kesiapan rilis OpsMind AI

Checkpoint: 29 Oktober 2026. Rilis fungsional yang sudah diverifikasi: [b6eaf5b, hasil PR #4](https://github.com/Tegar234-koplox/ops-mind-ai/pull/4). Tahap kesiapan rilis menambah pengujian kegagalan provider, dokumen dan atribusi paket; hasil CI untuk perubahan tersebut harus dilihat pada PR/head terkait.

## Bukti penerimaan

| Kemampuan | Hasil yang diamati | Bukti yang dapat diperiksa |
|---|---|---|
| Golden P0 | Lembur Rp1.380.000 vs Rp1.000.000 (+38%); output 830/1000 (-17%); 38 =31 hadir +3 cuti +4 pending; downtime union 110 menit | Seed dan tes integrasi; laporan Meta nyata tersimpan pada demo |
| Empat aturan P1 | ATT-001, OT-002, COST-001, COST-002 aktif pada demo; sumber belum lengkap tetap perlu verifikasi | [PR #2](https://github.com/Tegar234-koplox/ops-mind-ai/pull/2), tes PostgreSQL dan browser P1 |
| Agent Meta | Golden: dua putaran, tujuh tool/bukti, empat domain. Investigasi nyata ATT-001 dan COST-002 juga berhasil | Adapter/runner di internal/agent; laporan berlabel Meta pada demo, bukan hasil mock CI |
| Bukti dan review | Snapshot/hash, scope kasus, SSE reconnect, keputusan manusia serta audit tersimpan; review tidak menulis data bisnis | Tes integrasi/API; verifikasi daring dan setelah restart |
| P2 serah terima | Snapshot immutable dari fakta Go dan investigasi/review tersimpan, riwayat versi, telusuri kasus, cetak/ponsel | [PR #4](https://github.com/Tegar234-koplox/ops-mind-ai/pull/4), internal/handover, tes integrasi dan browser |
| Biaya P2 | Penyusunan dua laporan daring tidak menambah run/reservasi model; saldo sebelum/sesudah sama | Pemeriksaan API lokal privat; tes integrasi memastikan jalur tanpa model |
| UI hover/fokus | Navigasi, tab dan tautan bukti merespons hover; fokus keyboard dan viewport ponsel diperiksa | [PR #3](https://github.com/Tegar234-koplox/ops-mind-ai/pull/3), tes browser |
| Autentikasi/deployment | Render Free + Neon Free, HTTPS, API tanpa login 401, origin asing 403, cookie Secure/HttpOnly/SameSite Strict | [Demo terlindungi login](https://opsmind-ai-l0px.onrender.com); kredensial dibagikan pemilik secara privat |
| Regresi dan container rilis fungsional | Go vet/race/PostgreSQL, build/format, browser mock dan startup container lulus | [CI main b6eaf5b](https://github.com/Tegar234-koplox/ops-mind-ai/actions/runs/40904910986) |
| Gangguan provider | HTTP simulasi: retry sementara, 401, timeout, JSON/usage rusak, truncation/filter, redirect, budget/payload, tool/scope asing dan bukti kurang | internal/integration/agent_boundary_test.go; tes terarah lokal lulus, tanpa key/panggilan Meta nyata |

Laporan golden daring: case 6cf2a0b1f6e4a08521ec6619cb960c2c, run fc4ed857586151f2e08aed563cbf4dbe. P1: ATT-001 run 7c0e144ab0363c1934f3d28e1d966fb7; COST-002 run d47f049bf670286ed669d00cff8f7d83. Semua data sintetis. ID ini membantu pemilik memilih bukti yang tepat; hasil tambahan pengguna tidak ditimpa.

## Cara memeriksa

Ikuti [README](../README.md) untuk setup, seed dan pengujian. CI memakai mock eksplisit dan PostgreSQL 18; tes Meta nyata dijalankan terpisah. Integrasi tanpa TEST_DATABASE_URL di-skip, bukan dinyatakan lulus. Tes reset hanya boleh pada database berakhiran _test, tanpa server/worker lain yang memakai database tersebut.

Pengujian HTTP tambahan memakai server lokal httptest dan database opsmind_release_test. Panggilan sukses simulasi tetap disimpan dengan mode mock. Tes memeriksa jumlah retry, reservasi biaya, penolakan redirect dan tidak terjadinya HTTP call ketika budget/payload ditolak. Bukti satu domain menghasilkan insufficient_evidence dan rekomendasi dikosongkan. Ini menguji perilaku kode; bukan simulasi seluruh gangguan jaringan Render/Meta di dunia nyata.

Docker lokal belum menyediakan daemon yang dapat dipakai pada sesi pengembangan sebelumnya. Startup container benar-benar diperiksa CI Linux. Jangan menyatakan Docker lokal atau CI head baru lulus hanya berdasarkan Dockerfile. Salinan lisensi dalam kemasan runtime ditambahkan agar atribusi tetap tersedia bersama binary/frontend.

## Batas klaim

- Pending attendance bukan mangkir; cuti disetujui dikecualikan. Anomali biaya bukan bukti fraud atau efisiensi.
- Hipotesis Meta memerlukan penilaian manusia. JSON, hash dan referensi yang valid tidak membuktikan semua narasi benar.
- Snapshot serah terima bukan bukti penerimaan shift. Cetak/PDF memakai browser; tidak ada pengiriman otomatis.
- Render/Neon gratis dapat tidur dan memiliki kuota; tidak ada SLA atau janji proses 24/7.
- Kredensial, DSN, screenshot/trace dan JSON bukti lokal tidak dimasukkan ke repository publik.
- Procurement, warehouse, quality control, predictive maintenance, ERP, multi-fasilitas dan registrasi publik belum dikerjakan.

## Sisa pekerjaan pemilik

Video tiga menit, pemeriksaan Official Rules/Event Waiver final, pengaturan akses juri dan pengiriman Cerebral Valley belum dinyatakan selesai. Gunakan [checklist persyaratan submission](SUBMISSION_REQUIREMENTS.md) dan [rencana demo](HACKATHON_DEMO_PLAN.md). Penggunaan Codex tidak dianggap mendapat persetujuan eksplisit panitia. Tahap berikutnya tidak otomatis membuka modul bisnis baru.
