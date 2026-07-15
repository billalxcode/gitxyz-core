# Testing dan verifikasi

## Standar minimal

- Jalankan go test ./... setelah setiap perubahan yang memengaruhi build atau logika.
- Jika mengubah endpoint atau auth, uji manual lewat curl atau alat HTTP lain.
- Jika menambah fitur penting, prioritaskan unit/integration test bila memungkinkan.

## Workflow verifikasi

1. Jalankan gofmt jika ada perubahan Go.
2. Jalankan go test ./...
3. Jika ada fitur HTTP, uji endpoint dengan request yang relevan.
4. Pastikan response status, body, dan error handling sesuai ekspektasi.

## Kapan wajib verifikasi lebih lanjut

- Saat mengubah auth, JWT, middleware, atau route.
- Saat mengubah model/database.
- Saat mengubah fitur Git HTTP atau repository operations.
