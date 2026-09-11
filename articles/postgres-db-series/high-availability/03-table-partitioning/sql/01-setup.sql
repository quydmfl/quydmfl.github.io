-- Bảng events partition theo tháng (RANGE trên created_at), 6 tháng dữ liệu.

CREATE TABLE events (
    id SERIAL,
    created_at DATE NOT NULL,
    payload TEXT NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE events_2025_01 PARTITION OF events FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE events_2025_02 PARTITION OF events FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE events_2025_03 PARTITION OF events FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE events_2025_04 PARTITION OF events FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE events_2025_05 PARTITION OF events FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE events_2025_06 PARTITION OF events FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');

-- Bảng đối chứng KHÔNG partition, cùng dữ liệu, để so sánh
CREATE TABLE events_flat (
    id SERIAL,
    created_at DATE NOT NULL,
    payload TEXT NOT NULL
);

INSERT INTO events (created_at, payload)
SELECT date '2025-01-01' + (i % 181), 'payload ' || i
FROM generate_series(1, 600000) AS i;

INSERT INTO events_flat SELECT * FROM events;

-- Cùng 1 index trên cả 2 bảng để so sánh công bằng (index trên bảng cha tự tạo trên mọi partition).
CREATE INDEX events_created_at_idx ON events (created_at);
CREATE INDEX events_flat_created_at_idx ON events_flat (created_at);

VACUUM ANALYZE events;
VACUUM ANALYZE events_flat;

-- TODO (bài viết):
-- 1) EXPLAIN (ANALYZE, BUFFERS) SELECT count(*) FROM events
--    WHERE created_at >= '2025-03-01' AND created_at < '2025-04-01';
--    -- chỉ 1 partition (events_2025_03) xuất hiện trong plan
-- 2) So sánh với events_flat cho cùng điều kiện trên.
-- 3) EXPLAIN (ANALYZE, BUFFERS) SELECT count(*) FROM events WHERE payload LIKE 'payload 1%';
--    -- không lọc theo created_at -> quét hết mọi partition (Parallel Append)
-- 4) So sánh với events_flat cho cùng query ở bước 3.
