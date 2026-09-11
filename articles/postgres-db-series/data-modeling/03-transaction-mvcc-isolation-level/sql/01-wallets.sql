-- Bảng đơn giản để tái hiện lost update: 2 session cùng trừ số dư 1 ví.

CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    owner TEXT NOT NULL,
    balance NUMERIC(12, 2) NOT NULL CHECK (balance >= 0)
);

INSERT INTO wallets (owner, balance) VALUES ('alice', 100.00);

-- TODO (bài viết): kịch bản tái hiện lost update bằng 2 session psql song song:
--
-- Session A                              Session B
-- BEGIN;                                 BEGIN;
-- SELECT balance FROM wallets            SELECT balance FROM wallets
--   WHERE id = 1;  -- đọc 100              WHERE id = 1;  -- đọc 100
-- -- (chưa COMMIT, giả lập app tính toán ở code)
-- UPDATE wallets SET balance = 90        UPDATE wallets SET balance = 90
--   WHERE id = 1;                          WHERE id = 1;  -- block tới khi A COMMIT
-- COMMIT;                                COMMIT;
--
-- Kết quả: balance = 90 (đúng ra phải là 80 vì có 2 lần trừ 10) → 1 lần trừ bị mất.
--
-- Fix: thay SELECT thường bằng "SELECT ... FOR UPDATE" ở cả 2 session.
