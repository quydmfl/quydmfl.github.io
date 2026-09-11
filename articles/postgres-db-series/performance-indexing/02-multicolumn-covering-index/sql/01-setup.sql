CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    total NUMERIC(12, 2) NOT NULL
);

INSERT INTO orders (customer_id, status, total)
SELECT (1 + (random() * 5000)::int),
       (ARRAY['pending', 'paid', 'shipped'])[1 + (random() * 2)::int],
       (10000 + random() * 990000)::numeric(12, 2)
FROM generate_series(1, 300000);

CREATE INDEX orders_customer_status_idx ON orders (customer_id, status);
CREATE INDEX orders_customer_status_covering_idx ON orders (customer_id, status) INCLUDE (total);

ANALYZE orders;

-- TODO (bài viết):
-- 1) EXPLAIN (ANALYZE, BUFFERS) SELECT status FROM orders WHERE customer_id = 42;              -- dùng leftmost prefix
-- 2) EXPLAIN (ANALYZE, BUFFERS) SELECT status FROM orders WHERE status = 'paid';                -- KHÔNG dùng hiệu quả (bỏ qua cột đầu)
-- 3) EXPLAIN (ANALYZE, BUFFERS) SELECT customer_id, status, total FROM orders
--    WHERE customer_id = 42 AND status = 'paid';   -- so sánh Index Scan (index thường) vs Index Only Scan (index có INCLUDE)
