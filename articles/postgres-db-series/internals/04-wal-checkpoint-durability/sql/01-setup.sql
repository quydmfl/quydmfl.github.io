CREATE TABLE ledger (
    id SERIAL PRIMARY KEY,
    note TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- TODO (bài viết):
-- 1) Quan sát WAL hiện tại: SELECT pg_current_wal_lsn();
-- 2) List segment file thật: docker compose exec postgres ls -la /var/lib/postgresql/data/pg_wal/
-- 3) INSERT INTO ledger (note) VALUES ('trước khi kill'); rồi COMMIT (mặc định autocommit).
-- 4) Kill cứng container (không phải "docker compose down"):
--    docker compose kill -s SIGKILL postgres
--    docker compose start postgres
-- 5) Sau khi container khởi động lại, xác nhận dòng đã insert ở bước 3 vẫn còn:
--    SELECT * FROM ledger;
--    (Postgres tự động chạy crash recovery bằng WAL replay trước khi chấp nhận connection mới,
--    có thể thấy log "database system was not properly shut down" + "redo starts at ..."
--    trong "docker compose logs postgres".)
