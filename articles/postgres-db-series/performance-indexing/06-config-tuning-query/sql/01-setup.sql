CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    region TEXT NOT NULL,
    amount NUMERIC(12, 2) NOT NULL
);

INSERT INTO sales (region, amount)
SELECT (ARRAY['north', 'south', 'east', 'west'])[1 + (random() * 3)::int],
       (random() * 1000000)::numeric(12, 2)
FROM generate_series(1, 1000000);

CREATE INDEX sales_region_idx ON sales (region);

ANALYZE sales;

-- TODO (bài viết):
-- 1) SET max_parallel_workers_per_gather = 0;  -- cô lập biến số, tránh nhiễu bởi parallel worker
--    SET work_mem = '1MB';
--    EXPLAIN (ANALYZE) SELECT region, amount FROM sales ORDER BY amount;  -- "Sort Method: external merge Disk"
--    SET work_mem = '256MB';
--    EXPLAIN (ANALYZE) SELECT region, amount FROM sales ORDER BY amount;  -- "Sort Method: quicksort  Memory"
--    RESET max_parallel_workers_per_gather; RESET work_mem;
-- 2) SET random_page_cost = 4.0;  -- mặc định, giả định disk cơ
--    EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM sales WHERE region = 'south';  -- Bitmap Heap Scan
--    SET random_page_cost = 1.1;  -- hợp lý hơn cho SSD
--    EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM sales WHERE region = 'south';  -- Index Scan thẳng, nhanh hơn
--    RESET random_page_cost;
-- 3) CREATE TABLE sales_small AS SELECT * FROM sales LIMIT 100;
--    EXPLAIN (ANALYZE) SELECT sum(amount) FROM sales_small;  -- không parallel, bảng quá nhỏ
--    EXPLAIN (ANALYZE) SELECT sum(amount) FROM sales;        -- "Workers Planned: 2"
