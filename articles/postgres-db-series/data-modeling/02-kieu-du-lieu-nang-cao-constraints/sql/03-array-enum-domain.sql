-- Demo ARRAY (danh sách tag đơn giản), ENUM (tập giá trị cố định, có thứ tự tự nhiên),
-- và DOMAIN (tái sử dụng ràng buộc kiểu dữ liệu xuyên suốt nhiều bảng).

CREATE DOMAIN email AS TEXT
    CHECK (VALUE ~ '^[^@\s]+@[^@\s]+\.[^@\s]+$');

CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'cancelled');

CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}'
);

CREATE TABLE customers_v2 (
    id SERIAL PRIMARY KEY,
    contact_email email NOT NULL
);

CREATE TABLE orders_v2 (
    id SERIAL PRIMARY KEY,
    status order_status NOT NULL DEFAULT 'pending'
);

INSERT INTO articles (title, tags) VALUES
    ('Chuẩn hoá schema', ARRAY['postgresql', 'data-modeling']),
    ('EXCLUDE constraint', ARRAY['postgresql', 'constraints', 'booking']);

INSERT INTO customers_v2 (contact_email) VALUES ('a@example.com');

INSERT INTO orders_v2 (status) VALUES ('paid');

-- TODO (bài viết):
-- 1) Query bài viết có tag 'postgresql': SELECT * FROM articles WHERE 'postgresql' = ANY(tags);
-- 2) Insert email sai định dạng để thấy DOMAIN chặn ngay ở tầng DB:
--    INSERT INTO customers_v2 (contact_email) VALUES ('khong-phai-email');
-- 3) Sắp xếp theo ENUM order_status (paid < shipped vì thứ tự khai báo, không phải alphabet):
--    SELECT * FROM orders_v2 ORDER BY status;
