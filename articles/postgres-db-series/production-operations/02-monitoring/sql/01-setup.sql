CREATE EXTENSION IF NOT EXISTS pg_buffercache;

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    status TEXT NOT NULL,
    amount NUMERIC(12, 2) NOT NULL
);

INSERT INTO orders (status, amount)
SELECT (ARRAY['pending', 'paid', 'shipped'])[1 + (random() * 2)::int], random() * 100000
FROM generate_series(1, 50000);

CREATE VIEW ops_health AS
SELECT
  (SELECT round(100.0 * blks_hit / nullif(blks_hit + blks_read, 0), 2)
   FROM pg_stat_database WHERE datname = current_database()) AS cache_hit_ratio,
  (SELECT count(*) FROM pg_stat_activity) AS current_connections,
  (SELECT setting::int FROM pg_settings WHERE name = 'max_connections') AS max_connections,
  (SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock') AS waiting_on_lock,
  (SELECT coalesce(max(n_dead_tup), 0) FROM pg_stat_user_tables) AS max_dead_tup_any_table,
  (SELECT coalesce(max(age(relfrozenxid)), 0) FROM pg_class WHERE relkind = 'r') AS max_xid_age,
  (SELECT failed_count FROM pg_stat_archiver) AS archive_failed_count;

-- TODO (bài viết):
-- 1) SELECT * FROM ops_health;  -- baseline
-- 2) UPDATE orders SET amount = amount + 1;  rồi SELECT * FROM ops_health; -- max_dead_tup_any_table tăng
-- 3) Tái hiện lock contention (2 session) rồi SELECT * FROM ops_health;  -- waiting_on_lock = 1
-- 4) Drill-down ai đang chặn ai:
--    SELECT blocked.pid AS blocked_pid, blocked.query AS blocked_query,
--           blocking.pid AS blocking_pid, blocking.query AS blocking_query,
--           now() - blocking.query_start AS blocking_duration
--    FROM pg_stat_activity blocked
--    JOIN pg_locks bl ON bl.pid = blocked.pid AND NOT bl.granted
--    JOIN pg_locks kl ON kl.locktype = bl.locktype AND kl.database IS NOT DISTINCT FROM bl.database
--         AND kl.relation IS NOT DISTINCT FROM bl.relation AND kl.page IS NOT DISTINCT FROM bl.page
--         AND kl.tuple IS NOT DISTINCT FROM bl.tuple AND kl.pid <> bl.pid AND kl.granted
--    JOIN pg_stat_activity blocking ON blocking.pid = kl.pid;
