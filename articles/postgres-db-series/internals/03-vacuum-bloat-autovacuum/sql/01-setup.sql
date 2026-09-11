-- Bảng "counter" cập nhật liên tục để tạo dead tuple thật — mô phỏng pattern hay gây bloat
-- trong thực tế (counter/status cập nhật tần suất cao trên cùng 1 tập dòng nhỏ).

CREATE TABLE counters (
    id INTEGER PRIMARY KEY,
    value BIGINT NOT NULL DEFAULT 0
);

INSERT INTO counters (id, value) VALUES (1, 0);

-- TODO (bài viết):
-- 1) Đo kích thước ban đầu: SELECT pg_size_pretty(pg_relation_size('counters'));
-- 2) Tạo tải: chạy vòng lặp UPDATE counters SET value = value + 1 WHERE id = 1; vài nghìn lần
--    (dùng generate_series thay vì vòng lặp psql để nhanh):
--    UPDATE counters SET value = value + 1 WHERE id = 1;  -- lặp lại qua generate_series(1,5000)
-- 3) So sánh kích thước bảng sau khi tạo tải (trước khi VACUUM) với ban đầu.
-- 4) SELECT relname, n_dead_tup, n_live_tup, last_autovacuum FROM pg_stat_user_tables
--    WHERE relname = 'counters';
-- 5) VACUUM counters; rồi đo lại kích thước + n_dead_tup.
-- 6) age(relfrozenxid): SELECT relname, age(relfrozenxid) FROM pg_class WHERE relname = 'counters';
