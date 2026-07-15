# AGENTS

Panduan ini untuk agent yang bekerja di repo ini. Tujuan utamanya adalah menjaga konsistensi arsitektur, keamanan, dan pengalaman pengembangan untuk backend Go yang menggabungkan auth serta fitur Git hosting.

## Ringkasan repo

- Bahasa: Go
- Framework HTTP: Gin
- ORM: GORM
- Database: PostgreSQL
- Konfigurasi: Viper dengan file YAML
- Auth: JWT untuk route yang dilindungi

## Struktur utama

- [main.go](main.go): entry point aplikasi
- [internal/bootstrap](internal/bootstrap): inisialisasi database dan router
- [internal/api](internal/api): controller, service, middleware, DTO, dan route
- [internal/models](internal/models): model domain dan hooks GORM
- [internal/repository](internal/repository): akses data
- [modules/githttp](modules/githttp): logika Git HTTP dan protocol terkait

## Aturan kerja penting

1. Pertahankan arsitektur yang ada: controller -> service -> repository.
2. Pisahkan route publik dan route yang membutuhkan auth.
3. Gunakan middleware auth untuk route yang dilindungi.
4. Hindari perubahan besar yang mengganggu kompatibilitas API.
5. Selalu tangani error secara eksplisit dan jangan sembunyikan error.
6. Pertahankan validasi input di DTO dan service.
7. Saat mengubah model atau skema, pikirkan dampaknya pada GORM dan migrasi.
8. Verifikasi hasil perubahan dengan go test dan, bila perlu, uji endpoint lewat curl atau HTTP client.

## Dokumen terkait

- [docs/agents/architecture.md](docs/agents/architecture.md)
- [docs/agents/auth-and-routing.md](docs/agents/auth-and-routing.md)
- [docs/agents/database-and-models.md](docs/agents/database-and-models.md)
- [docs/agents/testing-and-verification.md](docs/agents/testing-and-verification.md)
