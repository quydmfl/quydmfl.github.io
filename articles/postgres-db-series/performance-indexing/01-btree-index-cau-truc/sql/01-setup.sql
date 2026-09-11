CREATE EXTENSION IF NOT EXISTS pageinspect;

CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 100.000 dòng: status 'done' chiếm 99%, status 'failed' chỉ chiếm ~1%
-- -> minh hoạ rõ khác biệt selectivity cho cùng 1 index.
INSERT INTO events (status, created_at)
SELECT CASE WHEN random() < 0.01 THEN 'failed' ELSE 'done' END,
       now() - (random() * interval '365 days')
FROM generate_series(1, 100000);

CREATE INDEX events_status_idx ON events (status);

ANALYZE events;

-- TODO (bài viết):
-- 1) EXPLAIN (ANALYZE) SELECT * FROM events WHERE status = 'failed';  -- selectivity thấp -> Index Scan
-- 2) EXPLAIN (ANALYZE) SELECT * FROM events WHERE status = 'done';    -- selectivity cao -> Seq Scan
-- 3) SELECT attname, n_distinct FROM pg_stats WHERE tablename = 'events' AND attname = 'status';
-- 4) bt_metap('events_status_idx'), bt_page_stats('events_status_idx', 1) -- soi cấu trúc B-tree thật
