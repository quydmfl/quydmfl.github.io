-- shard_b: khách hàng có tenant_id lẻ (ví dụ)
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    name TEXT NOT NULL
);

INSERT INTO customers (tenant_id, name) VALUES (1, 'Customer B1'), (3, 'Customer B3');
