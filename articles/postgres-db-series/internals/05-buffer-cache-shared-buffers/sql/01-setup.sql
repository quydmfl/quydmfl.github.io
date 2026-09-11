-- pg_buffercache: extension chính thức (contrib) để soi trực tiếp nội dung buffer pool.

CREATE EXTENSION IF NOT EXISTS pg_buffercache;

CREATE TABLE big_table (
    id SERIAL PRIMARY KEY,
    payload TEXT NOT NULL
);

INSERT INTO big_table (payload)
SELECT repeat('x', 200)
FROM generate_series(1, 200000);

-- TODO (bài viết):
-- 1) Reset thống kê để đo sạch: SELECT pg_stat_reset();
-- 2) Query lần đầu (cold, dữ liệu chưa nằm trong buffer pool nhiều):
--    EXPLAIN (ANALYZE, BUFFERS) SELECT count(*) FROM big_table WHERE payload LIKE 'x%';
--    -> so sánh "shared hit" vs "shared read" trong output.
-- 3) Chạy lại đúng câu trên lần 2 (warm): buffers hit tăng mạnh, read giảm.
-- 4) Hit ratio toàn database:
--    SELECT round(100.0 * blks_hit / nullif(blks_hit + blks_read, 0), 2) AS hit_ratio
--    FROM pg_stat_database WHERE datname = current_database();
-- 5) Soi buffer pool theo bảng cụ thể:
--    SELECT c.relname, count(*) AS buffers
--    FROM pg_buffercache b JOIN pg_class c ON b.relfilenode = pg_relation_filenode(c.oid)
--    WHERE c.relname = 'big_table' GROUP BY c.relname;
