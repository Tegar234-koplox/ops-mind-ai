# Atribusi pihak ketiga

Kode aplikasi OpsMind AI memakai [lisensi MIT](LICENSE). Lisensi tersebut tidak menggantikan hak cipta dan syarat komponen pihak ketiga. Salinan di bawah diambil dari versi dependensi yang terpasang pada 29 Oktober 2026; teks lisensi upstream dipertahankan.

| Komponen | Versi terkunci | Lisensi | Salinan lisensi |
|---|---|---|---|
| [Go standard library](https://go.dev/) | 1.27.3 | BSD-3-Clause | [Teks](licenses/go.txt) |
| [github.com/go-chi/chi/v5](https://github.com/go-chi/chi) | v5.5.0 | MIT | [Teks](licenses/go-chi.txt) |
| [github.com/golang-migrate/migrate/v4](https://github.com/golang-migrate/migrate) | v4.20.1 | MIT | [Teks](licenses/golang-migrate.txt) |
| [github.com/jackc/pgx/v5](https://github.com/jackc/pgx) | v5.11.0 | MIT | [Teks](licenses/pgx.txt) |
| [github.com/santhosh-tekuri/jsonschema/v6](https://github.com/santhosh-tekuri/jsonschema) | v6.0.3 | Apache-2.0 | [Teks](licenses/jsonschema.txt) |
| [golang.org/x/crypto](https://go.googlesource.com/crypto) | v0.58.0 | BSD-3-Clause | [Teks](licenses/x-crypto.txt) |
| [github.com/jackc/pgerrcode](https://github.com/jackc/pgerrcode) | v0.0.0-20220416144525-469b46aa5efa | MIT + PostgreSQL (data) | [Teks](licenses/pgerrcode.txt) |
| [github.com/jackc/pgpassfile](https://github.com/jackc/pgpassfile) | v1.0.0 | MIT | [Teks](licenses/pgpassfile.txt) |
| [github.com/jackc/pgservicefile](https://github.com/jackc/pgservicefile) | v0.0.0-20240606120523-5a60cdf6a761 | MIT | [Teks](licenses/pgservicefile.txt) |
| [github.com/jackc/puddle/v2](https://github.com/jackc/puddle) | v2.2.2 | MIT | [Teks](licenses/puddle.txt) |
| [golang.org/x/sync](https://go.googlesource.com/sync) | v0.24.0 | BSD-3-Clause | [Teks](licenses/x-sync.txt) |
| [golang.org/x/text](https://go.googlesource.com/text) | v0.43.0 | BSD-3-Clause | [Teks](licenses/x-text.txt) |
| [react](https://github.com/facebook/react) | 19.3.1 | MIT | [Teks](licenses/react.txt) |
| [react-dom](https://github.com/facebook/react) | 19.3.1 | MIT | [Teks](licenses/react-dom.txt) |
| [scheduler](https://github.com/facebook/react) | 0.28.0 | MIT | [Teks](licenses/scheduler.txt) |
| [@tanstack/react-query](https://github.com/TanStack/query) | 5.105.0 | MIT | [Teks](licenses/react-query.txt) |
| [@tanstack/query-core](https://github.com/TanStack/query) | 5.105.0 | MIT | [Teks](licenses/query-core.txt) |
| [lucide-react](https://github.com/lucide-icons/lucide) | 0.468.0 | ISC; atribusi Feather (MIT) | [Teks](licenses/lucide-react.txt) |
| [tailwindcss](https://github.com/tailwindlabs/tailwindcss) | 4.3.3 | MIT | [Teks](licenses/tailwindcss.txt) |

Inventaris ini mencakup modul dalam go.mod, pustaka standar Go, dependensi JavaScript runtime beserta scheduler/query-core, dan CSS Tailwind yang masuk hasil build. go.sum dan web/package-lock.json menyimpan resolusi dependensi; dependensi build/pengujian tetap memiliki lisensi masing-masing dalam paketnya. Perbarui inventaris dan salinannya saat versi komponen berubah.

Image runtime menyertakan LICENSE, dokumen ini, dan direktori licenses di /app. Image dasar Alpine serta paket sistem ca-certificates/tzdata mempunyai lisensi tersendiri; metadata paket tersedia melalui apk info --license di image. Inventaris aplikasi ini bukan audit lisensi seluruh sistem operasi.

Data manufaktur dibuat generator sintetis proyek. Tidak ada data pekerja nyata atau aset foto eksternal. Ikon berasal dari Lucide dengan atribusi di atas. Meta Model API adalah layanan eksternal dengan ketentuan layanan terpisah; model dan bobotnya tidak didistribusikan dalam repositori. Penggunaan Codex sebagai alat bantu pengembangan dijelaskan dalam [catatan persyaratan submission](docs/SUBMISSION_REQUIREMENTS.md).
