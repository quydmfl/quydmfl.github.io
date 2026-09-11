-- Schema cho demo window function + CTE:
-- - categories có parent_id (tự tham chiếu) để demo Recursive CTE (cây danh mục)
-- - products, sales để demo ranking / running total / LATERAL

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER REFERENCES categories(id),
    name TEXT NOT NULL
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    name TEXT NOT NULL,
    price NUMERIC(12, 2) NOT NULL
);

CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    sold_at DATE NOT NULL,
    quantity INTEGER NOT NULL,
    revenue NUMERIC(12, 2) NOT NULL
);
