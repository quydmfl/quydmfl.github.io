-- Seed: cây danh mục 2 tầng (dùng cho demo Recursive CTE), 30 sản phẩm chia đều 5 danh mục lá,
-- và ~6 tháng dữ liệu bán hàng (đủ để ranking/running total có ý nghĩa).

-- Cây danh mục
INSERT INTO categories (id, parent_id, name) VALUES
    (1, NULL, 'Electronics'),
    (2, 1, 'Laptops'),
    (3, 2, 'Gaming Laptops'),
    (4, 2, 'Ultrabooks'),
    (5, 1, 'Phones'),
    (6, NULL, 'Home & Kitchen'),
    (7, 6, 'Kitchen Appliances'),
    (8, 6, 'Furniture');
SELECT setval('categories_id_seq', 8);

-- 30 sản phẩm, chia đều cho 5 danh mục lá: Gaming Laptops(3), Ultrabooks(4), Phones(5),
-- Kitchen Appliances(7), Furniture(8) -- mỗi sản phẩm có 1 giá cố định
INSERT INTO products (category_id, name, price)
SELECT (ARRAY[3, 4, 5, 7, 8])[1 + (i % 5)],
       'Product ' || i,
       (50000 + (random() * 4950000))::numeric(12, 2)
FROM generate_series(1, 30) AS i;

-- ~6 tháng dữ liệu bán hàng: 1200 giao dịch rải ngẫu nhiên trên 30 sản phẩm và 180 ngày,
-- revenue = quantity * price. Sinh trực tiếp theo hàng (không qua LATERAL không tương quan)
-- để mỗi giao dịch có random() riêng biệt, tránh Postgres materialize và lặp lại cùng 1 tập giá trị.
INSERT INTO sales (product_id, sold_at, quantity, revenue)
SELECT p.id,
       date '2025-04-01' + (random() * 180)::int,
       g.quantity,
       g.quantity * p.price
FROM (
    SELECT (1 + (random() * 29)::int) AS product_id, (1 + (random() * 4)::int) AS quantity
    FROM generate_series(1, 1200)
) g
JOIN products p ON p.id = g.product_id;

ANALYZE;
