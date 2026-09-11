CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'replicator';

CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO events (message) VALUES ('dòng đầu tiên');

-- TODO (bài viết):
-- 1) Kill cứng primary: docker compose kill -s SIGKILL primary
-- 2) Promote replica: docker compose exec replica psql -U demo -d appdb -c "SELECT pg_promote();"
-- 3) Ghi thử vào replica (giờ đã là primary mới):
--    docker compose exec replica psql -U demo -d appdb -c "INSERT INTO events (message) VALUES ('sau failover');"
-- 4) Tái hiện split-brain: khởi động lại primary cũ TRƯỚC KHI báo cho nó biết đã có primary mới
--    -> cả 2 node cùng chấp nhận ghi độc lập.
-- 5) pg_rewind để đưa primary cũ quay lại làm replica (thay vì rebuild từ đầu):
--    docker compose stop primary   -- pg_rewind cần target server đã dừng hẳn
--    docker compose exec replica psql -U demo -d appdb -c "ALTER ROLE replicator WITH SUPERUSER;"
--      -- libpq method của pg_rewind cần đọc file nhị phân từ xa (pg_read_binary_file),
--      -- đòi hỏi quyền cao hơn REPLICATION thường
--    docker compose run --rm --entrypoint sh primary -c "
--      gosu postgres pg_rewind --target-pgdata=/var/lib/postgresql/data
--      --source-server='host=replica user=replicator password=replicator dbname=appdb' -P"
--    docker compose start primary
