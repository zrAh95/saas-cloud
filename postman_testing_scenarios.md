# POSTMAN DEBUGGING SCENARIOS - SaaS Cloud

## Base URL

```text
http://localhost:8080/api/v1
```

## Test Accounts

Pakai 3 akun biar flow permission gampang dicek:

- `owner@example.com`
- `shared@example.com`
- `random@example.com`

Password contoh:

```text
ValidPass123
```

## Environment Variables yang Disarankan

```text
BASE_URL=http://localhost:8080/api/v1
OWNER_ACCESS_TOKEN=
OWNER_REFRESH_TOKEN=
SHARED_ACCESS_TOKEN=
SHARED_REFRESH_TOKEN=
RANDOM_ACCESS_TOKEN=
RANDOM_REFRESH_TOKEN=
FILE_ID=
SHARE_ID=
FOLDER_ID=
CHILD_FILE_ID=
FOLDER_SHARE_ID=
```

## Suite 1 - Auth Register dan Verify

### 1. Register owner/shared/random

`POST {{BASE_URL}}/auth/register`

Body:

```json
{
  "email": "owner@example.com",
  "password": "ValidPass123"
}
```

Ulangi untuk:

- `shared@example.com`
- `random@example.com`

Expected:

- `200 OK`
- message: `Registrasi berhasil, silakan cek OTP`

### 2. Verify OTP untuk semua akun

`POST {{BASE_URL}}/auth/verify`

Body:

```json
{
  "email": "owner@example.com",
  "otp": "OTP_DARI_CONSOLE"
}
```

Expected:

- `200 OK`
- message: `Verifikasi berhasil`

## Suite 2 - Login dan JWT

### 3. Login owner

`POST {{BASE_URL}}/auth/login`

Body:

```json
{
  "email": "owner@example.com",
  "password": "ValidPass123"
}
```

Expected:

- `200 OK`
- dapat `access_token` dan `refresh_token`

Simpan ke:

- `OWNER_ACCESS_TOKEN`
- `OWNER_REFRESH_TOKEN`

### 4. Login shared dan random

Ulangi login yang sama untuk:

- `shared@example.com`
- `random@example.com`

Simpan token masing-masing.

### 5. Cek profile pakai access token

`GET {{BASE_URL}}/auth/profile`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- `200 OK`

### 6. Cek refresh token tidak boleh dipakai ke route protected

`GET {{BASE_URL}}/auth/profile`

Header:

```text
Authorization: Bearer {{OWNER_REFRESH_TOKEN}}
```

Expected:

- `401 Unauthorized`
- error: `Access token tidak valid`

### 7. Refresh token masih bisa dipakai untuk refresh

`POST {{BASE_URL}}/auth/refresh`

Body:

```json
{
  "refresh_token": "{{OWNER_REFRESH_TOKEN}}"
}
```

Expected:

- `200 OK`
- dapat access token baru

## Suite 3 - Upload dan File Owner

### 8. Owner upload file

`POST {{BASE_URL}}/files/upload`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Body:

- form-data
- key `file` tipe `File`

Expected:

- `200 OK`
- data file punya `id`

Simpan ke `FILE_ID`.

### 9. Owner lihat daftar file sendiri

`GET {{BASE_URL}}/files/`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- file yang baru diupload muncul

### 10. Owner buka file by id

`GET {{BASE_URL}}/files/{{FILE_ID}}`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- file kebuka / ter-stream

### 11. Owner download file

`GET {{BASE_URL}}/files/{{FILE_ID}}/download`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- response attachment download

## Suite 3A - Folder Logic

### 11A. Owner create root folder

`POST {{BASE_URL}}/files/folders`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
Content-Type: application/json
```

Body:

```json
{
  "name": "Project A"
}
```

Expected:

- `200 OK`
- data folder punya `id`
- `is_folder = true`

### 11B. Owner list root items

`GET {{BASE_URL}}/files/`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- folder `Project A` muncul di list root

### 11C. Owner upload file ke dalam folder

`POST {{BASE_URL}}/files/upload`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Body:

- form-data
- key `file` tipe `File`
- key `parent_id` isi folder id dari `Project A`

Expected:

- `200 OK`
- file tersimpan dengan `parent_id` folder tadi

### 11D. Owner lihat isi folder

`GET {{BASE_URL}}/files/folders/FOLDER_ID/contents`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- file yang baru diupload muncul di isi folder

### 11E. Owner hapus folder

`DELETE {{BASE_URL}}/files/FOLDER_ID`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- folder terhapus
- file anak di dalam folder ikut terhapus

### 11F. Owner buat subfolder di dalam folder

`POST {{BASE_URL}}/files/folders`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
Content-Type: application/json
```

Body:

```json
{
  "name": "Child Folder",
  "parent_id": {{FOLDER_ID}}
}
```

Expected:

- `200 OK`
- subfolder berhasil dibuat

### 11G. Owner upload child file ke dalam folder

`POST {{BASE_URL}}/files/upload`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Body:

- form-data
- key `file` tipe `File`
- key `parent_id` isi `{{FOLDER_ID}}`

Expected:

- `200 OK`
- response file punya `id`

Simpan ke `CHILD_FILE_ID`.

## Suite 4 - Share Flow

### 12. Owner share file ke shared user

`POST {{BASE_URL}}/files/{{FILE_ID}}/share`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Body:

```json
{
  "email": "shared@example.com"
}
```

Expected:

- `200 OK`
- `status` = `pending`
- ada `id` share row

Simpan ke `SHARE_ID`.

### 13. Prevent share ke diri sendiri

`POST {{BASE_URL}}/files/{{FILE_ID}}/share`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Body:

```json
{
  "email": "owner@example.com"
}
```

Expected:

- `400 Bad Request`
- error: `Tidak bisa share ke diri sendiri`

### 14. Prevent duplicate share

Ulangi request suite 12.

Expected:

- `400 Bad Request`
- error: `File sudah di-share ke user ini`

### 15. Shared user lihat incoming shares

`GET {{BASE_URL}}/files/shared`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- ada item dengan `share_id = {{SHARE_ID}}`
- status masih `pending`

## Suite 5 - Accept Share dan Permission

### 16. Shared user accept share

`POST {{BASE_URL}}/files/shares/{{SHARE_ID}}/accept`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- `status` berubah jadi `accepted`

### 17. Shared user cek list share lagi

`GET {{BASE_URL}}/files/shared`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- item yang sama sekarang `status = accepted`

### 18. Shared user buka file by id

`GET {{BASE_URL}}/files/{{FILE_ID}}`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`

### 19. Shared user download file

`GET {{BASE_URL}}/files/{{FILE_ID}}/download`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`

### 20. Random user tidak boleh akses file

`GET {{BASE_URL}}/files/{{FILE_ID}}`

Header:

```text
Authorization: Bearer {{RANDOM_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

### 21. Random user tidak boleh download file

`GET {{BASE_URL}}/files/{{FILE_ID}}/download`

Header:

```text
Authorization: Bearer {{RANDOM_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

### 22. Shared user tidak boleh accept share milik user lain

Login dengan `random@example.com`, lalu coba:

`POST {{BASE_URL}}/files/shares/{{SHARE_ID}}/accept`

Header:

```text
Authorization: Bearer {{RANDOM_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

## Suite 6 - Revoke Access

### 23. Owner revoke access shared user

`DELETE {{BASE_URL}}/files/{{FILE_ID}}/share`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
Content-Type: application/json
```

Body:

```json
{
  "email": "shared@example.com"
}
```

Expected:

- `200 OK`
- message: `Akses file berhasil dicabut`

### 24. Shared user tidak boleh akses file setelah revoke

`GET {{BASE_URL}}/files/{{FILE_ID}}`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

### 25. Shared user tidak boleh download file setelah revoke

`GET {{BASE_URL}}/files/{{FILE_ID}}/download`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

## Suite 7 - Delete File

### 26. Owner delete file

`DELETE {{BASE_URL}}/files/{{FILE_ID}}`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- `200 OK`

### 27. Owner cek file sudah hilang

`GET {{BASE_URL}}/files/{{FILE_ID}}`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Expected:

- `404 Not Found`

### 28. Shared user cek share list setelah file dihapus

`GET {{BASE_URL}}/files/shared`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- item untuk file itu sudah tidak ada

## Suite 8 - Folder Share Inheritance

### 29. Owner share folder ke shared user

`POST {{BASE_URL}}/files/{{FOLDER_ID}}/share`

Header:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
```

Body:

```json
{
  "email": "shared@example.com"
}
```

Expected:

- `200 OK`
- `status = pending`
- ada `id` share row

Simpan ke `FOLDER_SHARE_ID`.

### 30. Shared user lihat folder share yang masuk

`GET {{BASE_URL}}/files/shared`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- ada item folder dengan `share_id = {{FOLDER_SHARE_ID}}`
- `is_folder = true`

### 31. Shared user accept folder share

`POST {{BASE_URL}}/files/shares/{{FOLDER_SHARE_ID}}/accept`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- status berubah ke `accepted`

### 32. Shared user lihat folder di root list

`GET {{BASE_URL}}/files/`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- folder yang dishare muncul di root list

### 33. Shared user lihat isi shared folder

`GET {{BASE_URL}}/files/folders/{{FOLDER_ID}}/contents`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- child file / subfolder milik owner muncul

### 34. Shared user buka child file dari shared folder

`GET {{BASE_URL}}/files/{{CHILD_FILE_ID}}`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`
- child file bisa diakses walau tidak dishare langsung

### 35. Shared user download child file dari shared folder

`GET {{BASE_URL}}/files/{{CHILD_FILE_ID}}/download`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `200 OK`

### 36. Shared user upload file ke shared folder

`POST {{BASE_URL}}/files/upload`

Headers:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Body:

- form-data
- key `file` tipe `File`
- key `parent_id` isi `{{FOLDER_ID}}`

Expected:

- `200 OK`
- file baru berhasil dibuat di dalam shared folder

### 37. Shared user buat subfolder di shared folder

`POST {{BASE_URL}}/files/folders`

Headers:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
Content-Type: application/json
```

Body:

```json
{
  "name": "Shared Child Folder",
  "parent_id": {{FOLDER_ID}}
}
```

Expected:

- `200 OK`

### 38. Random user tidak boleh lihat isi shared folder

`GET {{BASE_URL}}/files/folders/{{FOLDER_ID}}/contents`

Header:

```text
Authorization: Bearer {{RANDOM_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

### 39. Owner revoke akses folder shared user

`DELETE {{BASE_URL}}/files/{{FOLDER_ID}}/share`

Headers:

```text
Authorization: Bearer {{OWNER_ACCESS_TOKEN}}
Content-Type: application/json
```

Body:

```json
{
  "email": "shared@example.com"
}
```

Expected:

- `200 OK`

### 40. Shared user tidak boleh akses child file setelah revoke folder

`GET {{BASE_URL}}/files/{{CHILD_FILE_ID}}`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

### 41. Shared user tidak boleh lihat isi folder setelah revoke

`GET {{BASE_URL}}/files/folders/{{FOLDER_ID}}/contents`

Header:

```text
Authorization: Bearer {{SHARED_ACCESS_TOKEN}}
```

Expected:

- `403 Forbidden`

## Debug Checklist Cepat

Kalau ada flow yang gagal, cek ini dulu:

- `.env` punya `JWT_SECRET_KEY`
- MySQL aktif dan tabel `tb_users`, `tb_otps`, `tb_files`, `tb_file_access` ada
- akun target share sudah register dan verify
- `Authorization` header formatnya `Bearer <token>`
- untuk accept share, pakai `SHARE_ID`, bukan `FILE_ID`
- untuk accept folder share, pakai `FOLDER_SHARE_ID`
- untuk revoke access, body JSON wajib kirim `email`
- file fisik benar-benar ada di folder `uploads/`
- kalau test shared folder, pastikan `CHILD_FILE_ID` memang anak dari `FOLDER_ID`

## Urutan Testing Paling Aman

1. Register dan verify 3 akun.
2. Login 3 akun dan simpan token.
3. Upload file sebagai owner.
4. Share file ke shared user.
5. Cek incoming share.
6. Accept share.
7. Test access shared user dan random user.
8. Revoke access.
9. Test access lagi setelah revoke.
10. Delete file.
11. Share folder dan test inheritance.
12. Revoke folder access dan test child access putus.
