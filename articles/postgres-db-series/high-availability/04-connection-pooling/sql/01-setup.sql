CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL
);

INSERT INTO events (message) VALUES ('demo connection pooling');

-- TODO (bài viết):
-- 1) Đẩy vượt max_connections=20 (trực tiếp vào Postgres, không qua PgBouncer):
--    dùng vài chục session "SELECT pg_sleep(30)" song song -> lỗi
--    "FATAL: sorry, too many clients already" ở connection vượt ngưỡng.
-- 2) Cùng tải đó nhưng qua PgBouncer (port 6432) -> không lỗi, PgBouncer tự xếp hàng/tái sử dụng.
-- 3) Test transaction mode phá SET session-level:
--    psql "postgresql://demo:demo@localhost:6432/appdb" -c "SET search_path TO public; SELECT current_setting('search_path');"
--    -- 2 câu lệnh có thể rơi vào 2 server connection khác nhau nếu ở giữa có transaction khác chen vào.
