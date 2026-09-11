CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    status TEXT NOT NULL,
    amount NUMERIC(12, 2) NOT NULL
);

INSERT INTO orders (status, amount)
SELECT (ARRAY['pending', 'paid', 'shipped'])[1 + (random() * 2)::int], random() * 100000
FROM generate_series(1, 100000);

-- TODO (bài viết) -- nâng cấp PG16 (image gốc) lên PG17 (cài thêm qua apk) trong CÙNG container,
-- vì pg_upgrade cần cả 2 bindir truy cập trực tiếp cùng filesystem. Chạy tay (không có script),
-- xem chi tiết đầy đủ trong bài viết:
--
-- 1) docker compose stop postgres   -- pg_upgrade cần old cluster đã dừng hẳn
-- 2) docker compose run --rm --entrypoint sh postgres -c "
--      chown postgres:postgres /var/lib/postgresql/data17
--      apk update && apk add --no-cache postgresql17 postgresql17-contrib
--      gosu postgres /usr/libexec/postgresql17/initdb -D /var/lib/postgresql/data17 -U demo --auth=trust
--    "
-- 3) pg_upgrade --check (thêm --username=demo vì POSTGRES_USER=demo, không phải 'postgres'):
--    docker compose run --rm --entrypoint sh postgres -c "
--      apk add --no-cache postgresql17-contrib
--      gosu postgres sh -c 'export LD_LIBRARY_PATH=/usr/lib; /usr/libexec/postgresql17/pg_upgrade
--        --old-bindir=/usr/local/bin --new-bindir=/usr/libexec/postgresql17
--        --old-datadir=/var/lib/postgresql/data --new-datadir=/var/lib/postgresql/data17
--        --username=demo --check'
--    "
-- 4) Bỏ --check để chạy upgrade thật (cùng câu lệnh).
-- 5) Khởi động cluster mới (port 5432, data17), EXPLAIN 1 query trước/sau khi chạy
--    vacuumdb --all --analyze-in-stages (lệnh pg_upgrade tự khuyến nghị ở cuối output).
