# Database dan model

## GORM

- Gunakan GORM untuk model dan query database.
- Model berada di internal/models.
- Hindari logika bisnis di model, kecuali hook kecil seperti hashing password atau inisialisasi field.
- Jika menambahkan hook BeforeCreate/BeforeUpdate, pastikan tidak mengganggu alur yang sudah ada.

## Praktik yang disukai

- Gunakan tag gorm yang sesuai untuk constraint, index, dan nullable field.
- Perhatikan nama kolom agar konsisten dengan field di database.
- Saat mengubah model, cek dampaknya pada repository dan response DTO.
- Jangan mengubah struktur model tanpa memikirkan migrasi dan kompatibilitas data.

## Konfigurasi database

- Konfigurasi database berasal dari Viper dan file YAML.
- Jaga agar koneksi database tetap aman dan tidak hardcode credential di source code.
