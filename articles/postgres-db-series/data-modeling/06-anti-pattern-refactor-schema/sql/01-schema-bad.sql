-- Schema "xấu" ban đầu — cố tình gộp đủ các anti-pattern liệt kê trong bài để refactor dần.
-- Các bước refactor nằm ở migrations/ (không tự chạy khi init container — chạy tay từng bước
-- bằng "docker compose exec postgres psql -U demo -d legacy_app -f migrations/0N-....sql").

CREATE TABLE users (
    id SERIAL PRIMARY KEY,               -- anti-pattern: SERIAL thay vì GENERATED ALWAYS AS IDENTITY
    email TEXT,                          -- anti-pattern: thiếu NOT NULL
    created_at TIMESTAMP                 -- anti-pattern: TIMESTAMP thay vì TIMESTAMPTZ
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) -- anti-pattern: thiếu index trên FK
);

-- anti-pattern: EAV thay vì cột thật cho thuộc tính user (locale, plan, v.v.)
CREATE TABLE user_attributes (
    user_id INTEGER NOT NULL REFERENCES users(id),
    key TEXT NOT NULL,
    value TEXT,
    PRIMARY KEY (user_id, key)
);
