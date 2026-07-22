# Arsitektur dan pola pengembangan

## Prinsip utama

- Gunakan pola layered yang sudah ada: controller, service, repository.
- Controller hanya menerima request dan mengembalikan response.
- Service berisi logika bisnis.
- Repository hanya berhubungan dengan database.
- Model domain sebaiknya tetap berada di internal/models.

## Pedoman implementasi

- Jangan menaruh logika database langsung di controller.
- Jangan menaruh logika bisnis di repository jika bisa dipindah ke service.
- Hindari membuat file besar yang menggabungkan banyak tanggung jawab.
- Saat menambahkan fitur, pilih lokasi yang konsisten dengan feature yang sudah ada.

## Struktur folder yang disarankan

- internal/api/controllers: handler HTTP
- internal/api/services: bisnis logic
- internal/api/dto: request/response DTO
- internal/api/middlewares: middleware Gin
- internal/api/routes: registrasi route
- internal/repository: akses data
- internal/models: model GORM
- modules/githttp: fitur Git protocol
