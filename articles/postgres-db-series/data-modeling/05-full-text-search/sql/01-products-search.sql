-- Demo full-text search (tsvector, GIN) + fuzzy search (pg_trgm) trên tên sản phẩm tiếng Việt.
-- Dùng text search config 'simple' vì PostgreSQL mặc định không có config riêng cho tiếng Việt
-- (không stem được tiếng Việt có dấu) — 'simple' chỉ tokenize + lowercase, không stem.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    search_vector TSVECTOR GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', name), 'A') ||
        setweight(to_tsvector('simple', description), 'B')
    ) STORED
);

CREATE INDEX products_search_vector_idx ON products USING GIN (search_vector);
CREATE INDEX products_name_trgm_idx ON products USING GIN (name gin_trgm_ops);

INSERT INTO products (name, description) VALUES
    ('Bàn phím cơ Akko 3068B', 'Bàn phím cơ hồng ngoại switch đỏ, kết nối không dây'),
    ('Bàn phím cơ Logitech G Pro', 'Bàn phím cơ thi đấu, switch clicky, kết nối có dây'),
    ('Chuột không dây Logitech MX Master', 'Chuột văn phòng cao cấp, pin sạc, kết nối bluetooth'),
    ('Chuột chơi game Razer DeathAdder', 'Chuột gaming có dây, cảm biến quang học độ chính xác cao'),
    ('Bàn ủi hơi nước Philips', 'Bàn ủi hơi nước công suất lớn, chống dính'),
    ('Bàn ủi khô Sunhouse', 'Bàn ủi khô giá rẻ, phù hợp gia đình nhỏ'),
    ('Nồi cơm điện Panasonic', 'Nồi cơm điện tử, lòng nồi chống dính, 1.8 lít'),
    ('Máy xay sinh tố Philips', 'Máy xay sinh tố công suất 600W, cối thuỷ tinh'),
    ('Laptop Dell XPS 13', 'Laptop mỏng nhẹ, màn hình cảm ứng, chip Intel i7'),
    ('Laptop gaming Asus ROG', 'Laptop chơi game, card đồ hoạ rời, tản nhiệt kép');

-- TODO (bài viết):
-- 1) Full-text: SELECT * FROM products WHERE search_vector @@ to_tsquery('simple', 'bàn & phím');
-- 2) Ranking: SELECT *, ts_rank(search_vector, query) FROM products, to_tsquery('simple', 'chuột') query
--    WHERE search_vector @@ query ORDER BY ts_rank(search_vector, query) DESC;
-- 3) Highlight: SELECT ts_headline('simple', name || ' - ' || description, to_tsquery('simple', 'khong & day'));
-- 4) Fuzzy (gõ sai dấu/nhầm chữ): SELECT name, similarity(name, 'ban fim akco')
--    FROM products ORDER BY similarity(name, 'ban fim akco') DESC LIMIT 5;
