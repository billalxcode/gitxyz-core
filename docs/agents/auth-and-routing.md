# Auth dan routing

## Routing

- Pisahkan route publik dan route yang membutuhkan auth.
- Route publik: register, login, forgot/reset password, verify email, dsb.
- Route terproteksi: logout, profile, endpoint yang butuh user context.
- Terapkan middleware auth sekali pada grup yang memang dilindungi.

## JWT

- Gunakan JWT untuk authenticated flow.
- Token harus dikirim lewat header Authorization dengan format Bearer <token>.
- Validasi token dilakukan di middleware.
- Jangan menaruh logika auth langsung di controller jika sudah ada middleware.

## Pedoman auth

- Password harus di-hash sebelum disimpan.
- Validasi input wajib dilakukan sebelum service dipanggil.
- Jangan mengembalikan detail internal error yang terlalu sensitif ke client.
- Saat login berhasil, kirim token dan data user yang relevan.
