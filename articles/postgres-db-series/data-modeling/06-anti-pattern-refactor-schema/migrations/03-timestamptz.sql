-- Bước 3: chuyển TIMESTAMP -> TIMESTAMPTZ. LƯU Ý: khác với NOT NULL ở bước 2, đổi kiểu cột
-- bằng ALTER COLUMN TYPE luôn REWRITE toàn bộ bảng và giữ ACCESS EXCLUSIVE LOCK suốt quá trình
-- -- không có "phiên bản NOT VALID" cho việc đổi kiểu. Với bảng lớn trên production, cách an
-- toàn hơn là: thêm cột mới -> backfill theo batch -> dual-write -> đổi tên cột (nằm ngoài phạm
-- vi demo này, bảng ở đây chỉ có vài dòng nên ALTER TYPE trực tiếp chấp nhận được).
--
-- AT TIME ZONE 'UTC' giả định dữ liệu TIMESTAMP hiện có đang được lưu theo giờ UTC (giả định
-- BẮT BUỘC phải xác nhận đúng với thực tế dữ liệu trước khi chạy migration này trên production).

BEGIN;
ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
COMMIT;
