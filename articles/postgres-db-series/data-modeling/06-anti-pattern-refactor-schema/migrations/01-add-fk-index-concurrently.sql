-- Bước 1: thêm index còn thiếu trên foreign key, không khoá bảng khi đang có traffic ghi.
-- CONCURRENTLY không thể chạy trong transaction block, nên chạy file này ngoài BEGIN/COMMIT.

CREATE INDEX CONCURRENTLY IF NOT EXISTS orders_user_id_idx ON orders (user_id);
