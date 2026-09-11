-- Bảng đơn giản để theo dõi xmin/xmax qua từng thao tác. Postgres cho phép SELECT trực tiếp
-- 2 cột hệ thống ẩn này mà không cần pageinspect (pageinspect chỉ cần khi muốn soi cả page).

CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner TEXT NOT NULL,
    balance NUMERIC(12, 2) NOT NULL
);

INSERT INTO accounts (owner, balance) VALUES ('alice', 100.00);

-- TODO (bài viết):
-- 1) SELECT xmin, xmax, * FROM accounts; -- xmin = xid của transaction vừa insert, xmax = 0
-- 2) UPDATE accounts SET balance = 90 WHERE owner = 'alice';
--    SELECT xmin, xmax, * FROM accounts; -- tuple mới xuất hiện, xmin mới; tuple cũ (nếu soi
--    bằng pageinspect) có xmax vừa được set, không còn hiện ra qua SELECT thường vì đã "chết"
--    với snapshot hiện tại.
-- 3) Chạy 2 session song song để so sánh xmin/xmax nhìn thấy được tại các thời điểm khác nhau.
