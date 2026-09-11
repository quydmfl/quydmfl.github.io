-- Bước 2: thêm NOT NULL cho users.email trên bảng "đang chạy production" mà không khoá bảng
-- để quét toàn bộ trong 1 transaction. Cách làm: CHECK ... NOT VALID (không khoá, không quét)
-- rồi VALIDATE CONSTRAINT (quét nhưng chỉ khoá share lock, không chặn ghi) ở 2 bước tách rời.

BEGIN;
ALTER TABLE users ADD CONSTRAINT users_email_not_null CHECK (email IS NOT NULL) NOT VALID;
COMMIT;

-- TODO (bài viết): trước khi VALIDATE, backfill các dòng email NULL còn sót lại (nếu có)
-- rồi mới chạy VALIDATE CONSTRAINT.

BEGIN;
ALTER TABLE users VALIDATE CONSTRAINT users_email_not_null;
COMMIT;

-- Sau khi đã validate và chắc chắn không còn dòng nào vi phạm, có thể chuyển hẳn sang NOT NULL
-- thật (PG 12+ dùng lại CHECK constraint đã validate nên bước này gần như tức thời):
-- ALTER TABLE users ALTER COLUMN email SET NOT NULL;
-- ALTER TABLE users DROP CONSTRAINT users_email_not_null;
