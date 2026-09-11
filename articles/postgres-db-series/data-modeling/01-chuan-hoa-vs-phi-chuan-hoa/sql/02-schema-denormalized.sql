-- Schema có denormalize CÓ CHỦ ĐÍCH: order_items lưu lại unit_price_at_order.
-- Đây là ví dụ denormalize *đúng* — bắt buộc về nghiệp vụ (giá sản phẩm đổi theo thời gian,
-- nhưng đơn hàng cũ phải giữ nguyên giá đã chốt), không phải chỉ để tối ưu tốc độ đọc.

CREATE TABLE order_items_v2 (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_at_order NUMERIC(12, 2) NOT NULL
);

-- TODO (bài viết): thêm cột tổng hợp orders.total_amount (denormalize để tối ưu đọc)
-- và benchmark EXPLAIN ANALYZE cho truy vấn "doanh thu theo sản phẩm theo tháng"
-- so với bản chuẩn hoá ở 01-schema-normalized.sql.
