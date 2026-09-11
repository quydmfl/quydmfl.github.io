-- Tạo role riêng cho replication (nguyên tắc: không dùng superuser cho kết nối replication).
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'replicator';

CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO events (message) VALUES ('dòng đầu tiên trước khi có replica');

-- TODO (bài viết):
-- 1) Xác nhận replica đã bắt kịp: (chạy ở container "replica")
--    docker compose exec replica psql -U demo -d appdb -c "SELECT * FROM events;"
-- 2) Insert thêm ở primary, xác nhận xuất hiện ở replica gần như tức thời.
-- 3) Thử ghi trực tiếp vào replica -> bị từ chối:
--    docker compose exec replica psql -U demo -d appdb -c "INSERT INTO events (message) VALUES ('x');"
--    ERROR: cannot execute INSERT in a read-only transaction
-- 4) Đo lag: SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn,
--    write_lag, flush_lag, replay_lag FROM pg_stat_replication;  -- chạy ở primary
