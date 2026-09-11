-- Bước 5: chuyển cột id từ SERIAL (default nextval trỏ tới 1 sequence rời) sang
-- GENERATED ALWAYS AS IDENTITY (chuẩn SQL, PG 10+). Phải DROP DEFAULT trước -- gọi thẳng
-- ADD GENERATED khi cột còn default cũ sẽ báo lỗi "column already has a default value".
--
-- START WITH phải lớn hơn giá trị id lớn nhất hiện có, tránh trùng id với dữ liệu cũ.
-- Xác định giá trị đúng bằng: SELECT max(id) + 1 FROM users;

ALTER TABLE users ALTER COLUMN id DROP DEFAULT;
ALTER TABLE users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (START WITH 6);

-- Sequence cũ (users_id_seq) giờ không còn được cột nào tham chiếu tới nữa -- dọn luôn:
-- DROP SEQUENCE IF EXISTS users_id_seq;
