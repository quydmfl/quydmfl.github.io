-- Demo GENERATED ALWAYS AS: cột tính toán tại tầng DB, luôn nhất quán, index được như cột thật.
-- Khác view/computed ở application: mọi client (kể cả 1 câu SQL ad-hoc) đều thấy cùng giá trị,
-- không phụ thuộc code application có tính đúng công thức hay không.

CREATE TABLE line_items (
    id SERIAL PRIMARY KEY,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12, 2) NOT NULL,
    total NUMERIC(14, 2) GENERATED ALWAYS AS (quantity * unit_price) STORED
);

INSERT INTO line_items (quantity, unit_price) VALUES
    (2, 500000),
    (3, 150000);

-- TODO (bài viết): thử insert/update trực tiếp cột total -> Postgres từ chối:
-- UPDATE line_items SET total = 999999 WHERE id = 1;
-- ERROR: column "total" can only be updated to DEFAULT
