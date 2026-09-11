-- city và country tương quan mạnh (mỗi city chỉ thuộc đúng 1 country) -- planner mặc định
-- coi 2 cột độc lập, dẫn tới ước lượng selectivity sai khi WHERE cả 2 cùng lúc.

CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    city TEXT NOT NULL,
    country TEXT NOT NULL
);

INSERT INTO addresses (city, country)
SELECT city, country FROM (
    SELECT (ARRAY['Hà Nội', 'Hồ Chí Minh', 'Đà Nẵng'])[1 + (random() * 2)::int] AS city,
           'Việt Nam' AS country
    FROM generate_series(1, 150000)
    UNION ALL
    SELECT (ARRAY['Bangkok', 'Chiang Mai'])[1 + (random() * 1)::int], 'Thái Lan'
    FROM generate_series(1, 50000)
) t;

ANALYZE addresses;

-- TODO (bài viết):
-- 1) EXPLAIN ANALYZE SELECT * FROM addresses WHERE city = 'Hà Nội' AND country = 'Việt Nam';
--    -- so sánh "rows=" ước lượng với "rows=... actual"
-- 2) CREATE STATISTICS addresses_city_country_stat (dependencies) ON city, country FROM addresses;
--    ANALYZE addresses;
--    -- chạy lại EXPLAIN ở bước 1, so sánh độ lệch ước lượng trước/sau
-- 3) Demo thống kê cũ (chạy trên bản mới, chưa làm bước 2 ở trên nếu muốn tái hiện đúng số liệu bài viết):
--    ALTER TABLE addresses SET (autovacuum_enabled = false);
--    EXPLAIN SELECT * FROM addresses WHERE country = 'Việt Nam';  -- baseline
--    INSERT INTO addresses (city, country)
--      SELECT 'Hồ Chí Minh', 'Việt Nam' FROM generate_series(1, 400000);
--    EXPLAIN ANALYZE SELECT * FROM addresses WHERE country = 'Việt Nam';  -- CHƯA ANALYZE
--    ANALYZE addresses;
--    EXPLAIN ANALYZE SELECT * FROM addresses WHERE country = 'Việt Nam';  -- ĐÃ ANALYZE
