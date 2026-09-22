# Supabase PostgreSQL

โปรเจกต์นี้ใช้ตารางเดียวสำหรับโจทย์สัมภาษณ์:

```text
public.app_users
├── id                  uuid, primary key
├── username            ชื่อที่ใช้แสดงผล
├── username_canonical  ชื่อสำหรับค้นหาและป้องกันชื่อซ้ำ
├── password_hash       bcrypt hash
└── created_at          วันเวลาที่สมัคร
```

## สร้างตาราง

Migration อยู่ที่:

```text
supabase/migrations/20260922000000_create_app_users.sql
```

รันด้วย Supabase CLI:

```bash
supabase link --project-ref <project-ref>
supabase db push
```

หรือคัดลอก SQL ใน migration ไปรันผ่าน Supabase SQL Editor

## เชื่อมต่อ backend

คัดลอก connection string จากปุ่ม **Connect** ใน Supabase Dashboard แล้วใส่ใน
`.env` พร้อมระบุ `sslmode`:

```text
DATABASE_URL=postgresql://postgres.<project-ref>:<password>@<pooler-host>:5432/postgres?sslmode=require
```

Migration เปิด RLS แต่ไม่สร้าง policy เพราะ frontend ของโจทย์นี้เรียกผ่าน Go API
เท่านั้น ไม่ได้อ่าน `app_users` ผ่าน Supabase client โดยตรง
