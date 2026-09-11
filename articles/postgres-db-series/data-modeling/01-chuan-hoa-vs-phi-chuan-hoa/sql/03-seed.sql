-- Seed dữ liệu để tái hiện đúng benchmark trong bài viết: 2000 khách hàng, 300 sản phẩm,
-- 50.000 đơn hàng trải trong 12 tháng, mỗi đơn 1-4 dòng (~200.000 order_items mỗi schema).
-- Chạy tự động khi container khởi động (docker-entrypoint-initdb.d), mất khoảng 5-10 giây.

INSERT INTO customers (name, email)
SELECT 'Customer ' || i, 'customer' || i || '@example.com'
FROM generate_series(1, 2000) AS i;

INSERT INTO products (name, price)
SELECT 'Product ' || i, (10000 + (random() * 1990000))::numeric(12, 2)
FROM generate_series(1, 300) AS i;

INSERT INTO orders (customer_id, created_at)
SELECT (1 + (random() * 1999)::int),
       timestamp '2025-09-01' + (random() * interval '365 days')
FROM generate_series(1, 50000) AS i;

-- order_items (schema chuẩn hoá): KHÔNG lưu giá, phải join products để biết giá
INSERT INTO order_items (order_id, product_id, quantity)
SELECT o.id,
       (1 + (random() * 299)::int),
       (1 + (random() * 4)::int)
FROM orders o
CROSS JOIN LATERAL generate_series(1, 1 + (random() * 3)::int) AS g;

-- order_items_v2 (schema denormalize): copy y hệt order_items, chốt thêm giá tại thời điểm đặt
INSERT INTO order_items_v2 (order_id, product_id, quantity, unit_price_at_order)
SELECT oi.order_id, oi.product_id, oi.quantity, p.price
FROM order_items oi
JOIN products p ON p.id = oi.product_id;

ANALYZE;
