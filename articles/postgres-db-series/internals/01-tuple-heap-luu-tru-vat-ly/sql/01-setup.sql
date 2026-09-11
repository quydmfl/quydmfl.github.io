-- pageinspect: extension chính thức của Postgres (contrib), cho phép soi trực tiếp nội dung
-- 1 page — dùng để chứng minh MVCC/tuple layout bằng dữ liệu thật thay vì chỉ mô tả lý thuyết.

CREATE EXTENSION IF NOT EXISTS pageinspect;

CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner TEXT NOT NULL,
    balance NUMERIC(12, 2) NOT NULL
);

INSERT INTO accounts (owner, balance) VALUES
    ('alice', 100.00),
    ('bob', 200.00);

-- TODO (bài viết):
-- 1) Soi tuple thật: SELECT * FROM heap_page_items(get_raw_page('accounts', 0));
-- 2) UPDATE accounts SET balance = 90 WHERE owner = 'alice'; rồi soi lại để thấy 2 tuple
--    (cũ + mới) cùng tồn tại trên page, khác xmin/xmax.
-- 3) TOAST: INSERT 1 dòng với giá trị TEXT rất lớn (repeat('x', 1000000)) vào 1 cột TEXT
--    riêng, rồi truy vấn pg_toast.pg_toast_<oid tương ứng> qua pg_class.reltoastrelid.
