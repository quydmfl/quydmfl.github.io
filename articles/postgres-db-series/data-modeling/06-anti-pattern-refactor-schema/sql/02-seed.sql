-- Seed dữ liệu "thật" cho legacy_app: vài dòng users có email NULL (để migration NOT NULL
-- ở bước 2 phải xử lý backfill trước khi validate), vài orders, và EAV attributes cho user.

INSERT INTO users (email, created_at) VALUES
    ('alice@example.com', '2025-01-10 08:00:00'),
    ('bob@example.com', '2025-02-15 09:30:00'),
    (NULL, '2025-03-01 10:00:00'),      -- dữ liệu bẩn có sẵn: email NULL
    ('dave@example.com', '2025-03-20 14:00:00'),
    (NULL, '2025-04-05 16:45:00');      -- dữ liệu bẩn có sẵn: email NULL

INSERT INTO orders (user_id) VALUES (1), (1), (2), (4);

INSERT INTO user_attributes (user_id, key, value) VALUES
    (1, 'locale', 'vi-VN'),
    (1, 'plan', 'pro'),
    (2, 'locale', 'en-US'),
    (2, 'plan', 'free'),
    (4, 'locale', 'vi-VN'),
    (4, 'plan', 'free');
