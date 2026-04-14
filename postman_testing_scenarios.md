# 🧪 POSTMAN TESTING SCENARIOS - SaaS Cloud Auth

## BASE URL
```
http://localhost:8080/api/v1
```

---

## 📋 TEST SUITE 1: REGISTER FLOW

### Test 1.1 - Invalid Email Format
**Endpoint:** `POST /auth/register`
**Status:** Should reject
```json
{
  "email": "not-an-email",
  "password": "ValidPass123"
}
```
**Expected Response:** 400
```json
{
  "error": "Format email tidak valid"
}
```

---

### Test 1.2 - Weak Password (Less than 8 chars)
**Endpoint:** `POST /auth/register`
```json
{
  "email": "test@example.com",
  "password": "Short1"
}
```
**Expected Response:** 400
```json
{
  "error": "Password tidak memenuhi kriteria: Password minimal 8 karakter"
}
```

---

### Test 1.3 - Password Missing Uppercase
**Endpoint:** `POST /auth/register`
```json
{
  "email": "test@example.com",
  "password": "lowercase123"
}
```
**Expected Response:** 400
```json
{
  "error": "Password tidak memenuhi kriteria: Password harus mengandung huruf besar"
}
```

---

### Test 1.4 - Password Missing Lowercase
**Endpoint:** `POST /auth/register`
```json
{
  "email": "test@example.com",
  "password": "UPPERCASE123"
}
```
**Expected Response:** 400
```json
{
  "error": "Password tidak memenuhi kriteria: Password harus mengandung huruf kecil"
}
```

---

### Test 1.5 - Password Missing Number
**Endpoint:** `POST /auth/register`
```json
{
  "email": "test@example.com",
  "password": "ValidPassword"
}
```
**Expected Response:** 400
```json
{
  "error": "Password tidak memenuhi kriteria: Password harus mengandung angka"
}
```

---

### Test 1.6 - Successful Register (NEW USER)
**Endpoint:** `POST /auth/register`
```json
{
  "email": "newuser@example.com",
  "password": "ValidPass123"
}
```
**Expected Response:** 200
```json
{
  "message": "Registrasi berhasil, silakan cek OTP",
  "data": null
}
```
**ACTION:** Copy OTP dari console log untuk Test 2.x

---

### Test 1.7 - Duplicate Email Registration
**Endpoint:** `POST /auth/register`
```json
{
  "email": "newuser@example.com",
  "password": "AnotherPass456"
}
```
**Expected Response:** 400
```json
{
  "error": "Email sudah terdaftar"
}
```

---

### Test 1.8 - Rate Limit Registration (6th attempt)
**Endpoint:** `POST /auth/register`
**Action:** Repeat Test 1.1-1.5 (5 times), then run this
```json
{
  "email": "test6@example.com",
  "password": "ValidPass123"
}
```
**Expected Response:** 429
```json
{
  "error": "Terlalu banyak upaya registrasi, coba lagi dalam beberapa menit"
}
```
**Note:** Tunggu 15 menit atau restart server buat reset

---

## 📋 TEST SUITE 2: OTP VERIFICATION

### Test 2.1 - Verify with Invalid OTP
**Endpoint:** `POST /auth/verify`
```json
{
  "email": "newuser@example.com",
  "otp": "000000"
}
```
**Expected Response:** 400
```json
{
  "error": "OTP tidak valid atau expired"
}
```

---

### Test 2.2 - Verify with Empty OTP
**Endpoint:** `POST /auth/verify`
```json
{
  "email": "newuser@example.com",
  "otp": ""
}
```
**Expected Response:** 400
```json
{
  "error": "Email & OTP wajib"
}
```

---

### Test 2.3 - Verify with Correct OTP ✅
**Endpoint:** `POST /auth/verify`
```json
{
  "email": "newuser@example.com",
  "otp": "PASTE_OTP_FROM_CONSOLE"
}
```
**Expected Response:** 200
```json
{
  "message": "Verifikasi berhasil",
  "data": null
}
```

---

### Test 2.4 - Verify OTP Rate Limit (11th attempt)
**Endpoint:** `POST /auth/verify`
**Action:** Run Test 2.1 ten times, then:
```json
{
  "email": "newuser@example.com",
  "otp": "000000"
}
```
**Expected Response:** 429
```json
{
  "error": "Terlalu banyak percobaan verifikasi OTP. Coba lagi dalam 15 menit"
}
```

---

## 📋 TEST SUITE 3: LOGIN FLOW

### Test 3.1 - Login with Invalid Email Format
**Endpoint:** `POST /auth/login`
```json
{
  "email": "not-email",
  "password": "ValidPass123"
}
```
**Expected Response:** 400
```json
{
  "error": "Format email tidak valid"
}
```

---

### Test 3.2 - Login with Wrong Password
**Endpoint:** `POST /auth/login`
```json
{
  "email": "newuser@example.com",
  "password": "WrongPassword123"
}
```
**Expected Response:** 401
```json
{
  "error": "Email atau password salah"
}
```

---

### Test 3.3 - Login with Non-Existent Email
**Endpoint:** `POST /auth/login`
```json
{
  "email": "nonexistent@example.com",
  "password": "ValidPass123"
}
```
**Expected Response:** 401
```json
{
  "error": "Email atau password salah"
}
```

---

### Test 3.4 - Successful Login ✅
**Endpoint:** `POST /auth/login`
```json
{
  "email": "newuser@example.com",
  "password": "ValidPass123"
}
```
**Expected Response:** 200
```json
{
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```
**ACTION:** Save tokens untuk Test 4.x dan Test 5.x

---

### Test 3.5 - Rate Limit Login (6th failed attempt)
**Endpoint:** `POST /auth/login`
**Action:** Run Test 3.2 five times, then:
```json
{
  "email": "newuser@example.com",
  "password": "WrongPassword123"
}
```
**Expected Response:** 429
```json
{
  "error": "Terlalu banyak percobaan login. Coba lagi dalam 15 menit. Sisa percobaan: 0"
}
```

---

## 📋 TEST SUITE 4: PROTECTED ENDPOINTS

### Test 4.1 - Access Protected Endpoint WITHOUT Token
**Endpoint:** `GET /auth/profile`
**Headers:** (None)
**Expected Response:** 401
```json
{
  "error": "Token kosong"
}
```

---

### Test 4.2 - Access with Invalid Token Format
**Endpoint:** `GET /auth/profile`
**Headers:**
```
Authorization: InvalidToken
```
**Expected Response:** 401
```json
{
  "error": "Format token salah"
}
```

---

### Test 4.3 - Access with Bearer but No Token
**Endpoint:** `GET /auth/profile`
**Headers:**
```
Authorization: Bearer
```
**Expected Response:** 401
```json
{
  "error": "Format token salah"
}
```

---

### Test 4.4 - Access with Valid Token ✅
**Endpoint:** `GET /auth/profile`
**Headers:**
```
Authorization: Bearer ACCESS_TOKEN_FROM_TEST_3.4
```
**Expected Response:** 200
```json
{
  "data": {
    "user_id": "1"
  }
}
```

---

## 📋 TEST SUITE 5: REFRESH TOKEN

### Test 5.1 - Refresh with Invalid Token
**Endpoint:** `POST /auth/refresh`
```json
{
  "refresh_token": "invalid.token.here"
}
```
**Expected Response:** 401
```json
{
  "error": "Token tidak valid"
}
```

---

### Test 5.2 - Refresh with Valid Refresh Token ✅
**Endpoint:** `POST /auth/refresh`
```json
{
  "refresh_token": "REFRESH_TOKEN_FROM_TEST_3.4"
}
```
**Expected Response:** 200
```json
{
  "message": "Refresh token berhasil",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```
**ACTION:** Save new access_token untuk Test 6.4

---

## 📋 TEST SUITE 6: LOGOUT & TOKEN BLACKLIST

### Test 6.1 - Logout WITHOUT Token
**Endpoint:** `POST /auth/logout`
**Headers:** (None)
**Expected Response:** 401
```json
{
  "error": "Token tidak valid"
}
```

---

### Test 6.2 - Logout with Valid Token ✅
**Endpoint:** `POST /auth/logout`
**Headers:**
```
Authorization: Bearer ACCESS_TOKEN_FROM_TEST_3.4
```
**Expected Response:** 200
```json
{
  "message": "Logout Berhasil",
  "data": null
}
```

---

### Test 6.3 - Access Protected After Logout (Blacklisted Token)
**Endpoint:** `GET /auth/profile`
**Headers:**
```
Authorization: Bearer ACCESS_TOKEN_FROM_TEST_6.2
```
**Expected Response:** 401
```json
{
  "error": "Token sudah logout"
}
```

---

## 🎯 QUICK TEST CHECKLIST

| # | Scenario | Endpoint | Expected Status |
|---|----------|----------|-----------------|
| 1.1 | Invalid email | POST /auth/register | 400 |
| 1.6 | Successful register | POST /auth/register | 200 |
| 1.7 | Duplicate email | POST /auth/register | 400 |
| 2.3 | Correct OTP | POST /auth/verify | 200 |
| 3.4 | Successful login | POST /auth/login | 200 |
| 4.4 | Protected endpoint | GET /auth/profile | 200 |
| 5.2 | Refresh token | POST /auth/refresh | 200 |
| 6.2 | Logout | POST /auth/logout | 200 |
| 6.3 | Blacklist check | GET /auth/profile | 401 |

---

## 💡 TIPS BESOK

1. **Buat folder di Postman:** Collection "SaaS-Cloud" → folder per test suite
2. **Environment variables:** Buat `JWT_TOKEN` dan `REFRESH_TOKEN` di Postman untuk auto-save
3. **Pre-request script:** Auto-extract token dari response terus assign ke environment
4. **Console log:** Ctrl+Alt+C buat lihat request/response detail
5. **Rate limit reset:** Restart server buat reset 15 menit window

---

## 🔐 IMPORTANT

- **OTP:** Check console log saat register, copy-paste ke verify
- **Rate limit:** Setelah kena limit, tunggu 15 menit atau restart
- **Token expiry:** Access token 15 menit, refresh token 7 hari
- **Logout persist:** Server restart akan reset blacklist (ini known limitation)
