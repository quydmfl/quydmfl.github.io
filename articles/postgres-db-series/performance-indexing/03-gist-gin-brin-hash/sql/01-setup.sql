CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    message TEXT NOT NULL
);

-- created_at tăng dần đúng theo thứ tự insert -- điều kiện lý tưởng cho BRIN
INSERT INTO logs (created_at, message)
SELECT timestamp '2025-01-01' + (i || ' seconds')::interval, 'log message ' || i
FROM generate_series(1, 1000000) AS i;

CREATE INDEX logs_created_at_btree_idx ON logs USING btree (created_at);
CREATE INDEX logs_created_at_brin_idx ON logs USING brin (created_at);
CREATE INDEX logs_id_hash_idx ON logs USING hash (id);

VACUUM ANALYZE logs;

-- Dữ liệu created_at trải từ 2025-01-01 tới khoảng 2025-01-12 (1 giây/dòng, 1 triệu dòng).
-- TODO (bài viết):
-- 1) So sánh kích thước:
--    SELECT pg_size_pretty(pg_relation_size('logs_created_at_btree_idx')),
--           pg_size_pretty(pg_relation_size('logs_created_at_brin_idx'));
-- 2) EXPLAIN (ANALYZE, BUFFERS) SELECT count(*) FROM logs
--    WHERE created_at BETWEEN '2025-01-05' AND '2025-01-06';
--    -- DROP từng index (btree/brin) luân phiên để so sánh plan dùng riêng biệt từng loại.
-- 3) EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM logs WHERE id = 500000;  -- Hash index cho equality
-- 4) EXPLAIN SELECT * FROM logs WHERE id > 500000 AND id < 500010;    -- Hash KHÔNG hỗ trợ range
