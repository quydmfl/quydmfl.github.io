CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner TEXT NOT NULL,
    balance NUMERIC(12, 2) NOT NULL
);

INSERT INTO accounts (owner, balance) VALUES ('alice', 1000.00);

-- TODO (bài viết):
-- 1) pg_dump/pg_restore:
--    docker compose exec postgres pg_dump -U demo -d appdb -F c -f /tmp/appdb.dump
--    docker compose exec postgres psql -U demo -d appdb -c "CREATE DATABASE appdb_restored;"
--    docker compose exec postgres pg_restore -U demo -d appdb_restored /tmp/appdb.dump
--
-- 2) PITR đầy đủ (chạy tay trong container, không có sẵn script — xem chi tiết trong bài viết):
--    a. Base backup (T0):
--       docker compose exec postgres gosu postgres pg_basebackup -U demo -D /var/lib/postgresql/basebackup -F p -X stream -P
--    b. Ghi thêm dữ liệu (T1): INSERT INTO accounts (owner, balance) VALUES ('bob', 500.00);
--    c. DROP TABLE nhầm (T2): DROP TABLE accounts;
--    d. Ghi thêm sau khi đã lỡ tay (T3): tạo bảng/dữ liệu khác
--    e. SELECT pg_switch_wal();  -- ép đóng segment hiện tại để nó được archive ngay,
--       nếu không PITR sẽ báo "recovery ended before configured recovery target was reached"
--    f. Copy basebackup sang thư mục restore, thêm recovery.signal + postgresql.auto.conf:
--       restore_command = 'cp /var/lib/postgresql/wal_archive/%f %p'
--       recovery_target_time = '<thời điểm giữa T1 và T2>'
--       recovery_target_action = 'promote'
--       port = 5433
--    g. gosu postgres pg_ctl start -D /var/lib/postgresql/restored
--    h. Xác nhận: accounts còn dữ liệu T1, bảng tạo ở T3 không tồn tại.
