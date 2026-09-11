-- Demo JSONB cho thuộc tính động (product attributes): mỗi loại sản phẩm có tập thuộc tính
-- khác nhau (áo có "size"/"color", laptop có "ram"/"cpu") — không hợp lý để mỗi thuộc tính
-- là 1 cột riêng trong bảng products (sẽ có hàng trăm cột NULL).

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    attributes JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX products_attributes_idx ON products USING GIN (attributes);

INSERT INTO products (name, attributes) VALUES
    ('Áo thun basic', '{"size": "L", "color": "black"}'),
    ('Laptop Dell XPS', '{"ram_gb": 16, "cpu": "i7-1360P"}'),
    ('Áo thun basic', '{"size": "M", "color": "white"}');

-- TODO (bài viết):
-- 1) Query theo 1 field trong JSONB (dùng được GIN index ở trên):
--    SELECT * FROM products WHERE attributes @> '{"size": "L"}';
-- 2) Query theo key tồn tại: SELECT * FROM products WHERE attributes ? 'ram_gb';
