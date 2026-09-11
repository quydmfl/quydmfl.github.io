-- Bước 4: refactor EAV (user_attributes) thành cột thật trên users, cho 2 thuộc tính cố định
-- và luôn có mặt (locale, plan). EAV chỉ còn hợp lý cho thuộc tính THỰC SỰ động/hiếm khi query
-- theo điều kiện -- ở đây locale/plan bị lạm dụng EAV cho thứ đáng ra là cột thật.

BEGIN;

ALTER TABLE users
    ADD COLUMN locale TEXT,
    ADD COLUMN plan TEXT NOT NULL DEFAULT 'free';

UPDATE users u
SET locale = a.value
FROM user_attributes a
WHERE a.user_id = u.id AND a.key = 'locale';

UPDATE users u
SET plan = a.value
FROM user_attributes a
WHERE a.user_id = u.id AND a.key = 'plan';

COMMIT;

-- TODO (bài viết): sau khi xác nhận ứng dụng đã đọc/ghi qua cột mới (dual-write một thời gian
-- trên production thật), mới xoá dữ liệu EAV cũ:
-- DELETE FROM user_attributes WHERE key IN ('locale', 'plan');
