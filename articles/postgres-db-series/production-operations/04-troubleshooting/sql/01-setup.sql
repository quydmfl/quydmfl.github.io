CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner TEXT NOT NULL,
    balance NUMERIC(12, 2) NOT NULL
);

INSERT INTO accounts (owner, balance) VALUES ('alice', 100.00), ('bob', 200.00);

CREATE TABLE orders (id SERIAL PRIMARY KEY, customer_id INTEGER, total NUMERIC);
INSERT INTO orders (customer_id, total)
SELECT (random() * 1000)::int, random() * 1000
FROM generate_series(1, 100000);

-- TODO (bài viết):
-- 1) Lock contention: session A "BEGIN; UPDATE accounts SET balance = 90 WHERE id = 1; SELECT pg_sleep(30);"
--    (không commit), session B "UPDATE accounts SET balance = 80 WHERE id = 1;" -> B bị treo.
--    Chẩn đoán: SELECT pid, state, wait_event_type, now() - query_start, query FROM pg_stat_activity
--    WHERE datname = current_database() AND pid <> pg_backend_pid();
--    Giải quyết: SELECT pg_terminate_backend(<pid session A>);  -- hoặc pg_cancel_backend cho trường hợp
--    chỉ muốn huỷ query, giữ session (test với 1 câu SELECT pg_sleep(30) đơn giản, không BEGIN).
--
-- 2) Runaway query: SELECT count(*) FROM orders o1, orders o2; (cartesian product cố ý).
--    Phát hiện: SELECT pid, now() - query_start, query FROM pg_stat_activity WHERE state = 'active';
--    Ngăn tái diễn: SET statement_timeout = '2s'; rồi chạy lại -> tự huỷ, không cần người can thiệp.
--
-- 3) Disk full: KHÔNG chạy trên container docker-compose này (dữ liệu chung) -- dùng container
--    riêng với tmpfs giới hạn dung lượng để mô phỏng an toàn:
--    docker run -d --name pg-diskfull-test \
--      --tmpfs /var/lib/postgresql/data:size=150m,uid=70,gid=70 \
--      -e POSTGRES_USER=demo -e POSTGRES_PASSWORD=demo -e POSTGRES_DB=appdb postgres:16-alpine
--    Rồi tạo tải ghi lớn tới khi gặp "No space left on device", quan sát Postgres vẫn phục vụ
--    được read, và DROP bớt dữ liệu không cần thiết để lấy lại chỗ trống, ghi thử lại thành công.
--    docker rm -f pg-diskfull-test  -- dọn dẹp sau khi xong
