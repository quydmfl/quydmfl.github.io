CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    status TEXT NOT NULL,
    email TEXT NOT NULL
);

-- 'pending' chỉ chiếm ~2% -- điều kiện lý tưởng cho partial index
INSERT INTO orders (status, email)
SELECT CASE WHEN random() < 0.02 THEN 'pending'
            WHEN random() < 0.5 THEN 'done'
            ELSE 'cancelled' END,
       'User' || i || '@Example.com'
FROM generate_series(1, 300000) AS i;

ANALYZE orders;

-- TODO (bài viết):
-- 1) CREATE INDEX orders_status_full_idx ON orders (status);
--    CREATE INDEX orders_status_pending_idx ON orders (status) WHERE status = 'pending';
--    -- so sánh pg_relation_size 2 index
-- 2) CREATE INDEX orders_email_lower_idx ON orders (lower(email));
--    EXPLAIN SELECT * FROM orders WHERE lower(email) = 'user123@example.com';  -- dùng được index
--    EXPLAIN SELECT * FROM orders WHERE email = 'User123@Example.com';         -- KHÔNG dùng index trên (email) thô nếu chưa tạo
