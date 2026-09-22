# Backend

Backend นี้ตั้งใจให้เป็น layered architecture แบบง่ายสำหรับอธิบายในสัมภาษณ์:

```text
HTTP request
    ↓
httpapi       รับ/ตอบ JSON และเลือก HTTP status
    ↓
service       ทำ Register, Login, CurrentUser
    ↓
repository    อ่านและเขียน PostgreSQL
```

ส่วนประกอบที่เหลือ:

- `domain` — User และกฎ validate
- `infrastructure/security` — bcrypt
- `infrastructure/token` — JWT
- `config` — อ่าน `DATABASE_URL` และ `JWT_SECRET`
- `cmd/api/main.go` — ประกอบทุกส่วนและเปิด HTTP server

## อธิบาย flow ในสัมภาษณ์

ตัวอย่าง Login:

1. Handler อ่าน `username` และ `password` จาก JSON
2. Service ขอ User จาก Repository
3. Service ใช้ bcrypt ตรวจ password
4. ถ้าถูกต้อง Service สร้าง JWT
5. Handler ตอบ token กลับเป็น JSON

Interface มีเฉพาะจุดที่ service ต้องพึ่งพาของภายนอก ได้แก่ repository, password
hasher และ token issuer จึงเขียน unit test ด้วย fake ได้โดยไม่ต้องต่อ Supabase

## รัน

```bash
cp .env.example .env
# กรอก DATABASE_URL และ JWT_SECRET
make backend
```

รายละเอียดสร้างตาราง Supabase ดูที่ `DATABASE.md`

## ทดสอบ

```bash
make test-backend
```

Test แบ่งตามสิ่งที่ต้องการพิสูจน์:

- `domain` — validation
- `service` — business flow ด้วย fake dependencies
- `repository/postgres` — SQL และ error mapping
- `httpapi` — endpoint ตั้งแต่ request ถึง response
- `security` / `token` — bcrypt และ JWT
