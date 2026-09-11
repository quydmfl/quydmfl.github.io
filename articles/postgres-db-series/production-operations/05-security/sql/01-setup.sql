-- Bảng multi-tenant để demo Row-Level Security. tenant khớp CHÍNH XÁC tên role
-- (dùng cho policy "tenant = current_user") -- đây là điều kiện bắt buộc, lệch tên là chính sách
-- sẽ lọc sai (0 dòng), không có cảnh báo nào khác.
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    tenant TEXT NOT NULL,
    title TEXT NOT NULL
);

INSERT INTO documents (tenant, title) VALUES
    ('app_tenant_a', 'Tài liệu A1'),
    ('app_tenant_a', 'Tài liệu A2'),
    ('app_tenant_b', 'Tài liệu B1');

CREATE ROLE app_tenant_a LOGIN PASSWORD 'app_tenant_a';
GRANT SELECT ON documents TO app_tenant_a;

-- Role chủ bảng KHÔNG phải superuser, dùng để demo FORCE ROW LEVEL SECURITY (khác superuser,
-- superuser luôn bypass RLS vô điều kiện, FORCE không có tác dụng với superuser).
CREATE ROLE app_owner LOGIN PASSWORD 'app_owner';

-- TODO (bài viết):
-- 1) psql "postgresql://app_tenant_a:app_tenant_a@localhost:5432/appdb" -c "SELECT * FROM documents;"
--    -- TRƯỚC khi bật RLS: thấy hết cả 3 dòng (kể cả app_tenant_b)
-- 2) ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
--    CREATE POLICY tenant_isolation ON documents USING (tenant = current_user);
--    Chạy lại bước 1 -> chỉ còn 2 dòng của app_tenant_a.
-- 3) psql -U demo -d appdb -c "SELECT * FROM documents;"  -- superuser vẫn thấy cả 3 (bypass RLS)
-- 4) ALTER TABLE documents OWNER TO app_owner;
--    ALTER TABLE documents FORCE ROW LEVEL SECURITY;
--    psql "postgresql://app_owner:app_owner@localhost:5432/appdb" -c "SELECT * FROM documents;"
--    -- 0 dòng: app_owner không phải superuser, FORCE buộc áp dụng policy kể cả với chủ bảng.
