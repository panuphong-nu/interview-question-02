-- User table for the interview assignment.
-- Passwords are stored as bcrypt hashes by the Go backend.

create table if not exists public.app_users (
    id uuid primary key default gen_random_uuid(),
    username varchar(32) not null,
    username_canonical varchar(32) not null,
    password_hash text not null,
    created_at timestamptz not null default now(),

    constraint app_users_username_canonical_key
        unique (username_canonical)
);

-- The frontend talks to the Go API, not directly to this table. With RLS
-- enabled and no policies, anon/authenticated Supabase clients cannot read it.
alter table public.app_users enable row level security;
