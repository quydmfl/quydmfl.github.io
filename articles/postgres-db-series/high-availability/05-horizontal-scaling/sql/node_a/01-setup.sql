-- shard_a: khách hàng có tenant_id chẵn (ví dụ)
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    name TEXT NOT NULL
);

INSERT INTO customers (tenant_id, name) VALUES (2, 'Customer A2'), (4, 'Customer A4');

-- TODO (bài viết) -- chạy tay SAU KHI cả node_a và node_b đã lên (không đưa vào init script
-- vì lúc node_a khởi tạo, node_b có thể chưa sẵn sàng để kết nối):
--
-- CREATE EXTENSION postgres_fdw;
-- CREATE SERVER shard_b FOREIGN DATA WRAPPER postgres_fdw
--   OPTIONS (host 'node_b', port '5432', dbname 'shard_b');
-- CREATE USER MAPPING FOR demo SERVER shard_b OPTIONS (user 'demo', password 'demo');
-- CREATE SCHEMA shard_b_schema;
-- IMPORT FOREIGN SCHEMA public FROM SERVER shard_b INTO shard_b_schema;
--
-- Rồi query gộp cả 2 shard:
-- SELECT 'node_a' AS shard, * FROM customers
-- UNION ALL
-- SELECT 'node_b' AS shard, * FROM shard_b_schema.customers
-- ORDER BY tenant_id;
