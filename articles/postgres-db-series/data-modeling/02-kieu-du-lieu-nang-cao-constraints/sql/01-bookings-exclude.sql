-- Demo EXCLUDE constraint: chặn double-booking (trùng phòng + trùng khung giờ) ngay ở tầng DB.
-- btree_gist cần thiết để kết hợp cột thường (room_id, dùng "=") với range (during, dùng "&&")
-- trong cùng một EXCLUDE constraint.

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL,
    during TSTZRANGE NOT NULL,
    EXCLUDE USING gist (room_id WITH =, during WITH &&)
);

INSERT INTO bookings (room_id, during) VALUES
    (1, '[2026-10-01 09:00+07, 2026-10-01 10:00+07)');

-- TODO (bài viết): minh hoạ insert dòng dưới đây thất bại vì trùng phòng + trùng giờ:
-- INSERT INTO bookings (room_id, during)
-- VALUES (1, '[2026-10-01 09:30+07, 2026-10-01 10:30+07)');
