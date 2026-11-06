-- Seed data.
-- NOTE: IDs are deterministic because this seed runs on a fresh database.
-- Product `stock` values are the FINAL state after the two sample orders below.

INSERT INTO products (sku, name, price, stock, is_active, created_at) VALUES
('SKU-0001', 'Beras Premium 5kg',    75000,  38, TRUE,  now() - interval '20 days'),
('SKU-0002', 'Minyak Goreng 2L',     32000,   8, TRUE,  now() - interval '20 days'),
('SKU-0003', 'Gula Pasir 1kg',       16500,  56, TRUE,  now() - interval '20 days'),
('SKU-0004', 'Telur Ayam 1kg',       28000,   5, TRUE,  now() - interval '20 days'),
('SKU-0005', 'Tepung Terigu 1kg',    12000,  75, TRUE,  now() - interval '20 days'),
('SKU-0006', 'Kopi Bubuk 250g',      45000,  24, TRUE,  now() - interval '20 days'),
('SKU-0007', 'Teh Celup 100s',       18500,   0, TRUE,  now() - interval '20 days'),
('SKU-0008', 'Susu UHT 1L',          19500,  30, TRUE,  now() - interval '20 days'),
('SKU-0009', 'Sabun Mandi 85g',       5500, 120, TRUE,  now() - interval '20 days'),
('SKU-0010', 'Shampoo 170ml',        23000,   9, TRUE,  now() - interval '20 days'),
('SKU-0011', 'Deterjen 800g',        21000,  45, TRUE,  now() - interval '20 days'),
('SKU-0012', 'Pasta Gigi 190g',      14000,   0, TRUE,  now() - interval '20 days'),
('SKU-0013', 'Air Mineral 600ml',     4000, 190, TRUE,  now() - interval '20 days'),
('SKU-0014', 'Mie Instan Goreng',     3500, 150, TRUE,  now() - interval '20 days'),
('SKU-0015', 'Produk Lama Discontinued', 10000, 0, FALSE, now() - interval '20 days');

-- Stock history: initial "in" for every product that started with stock.
-- (SKU-0007, SKU-0012, SKU-0015 start at zero and have no movement history.)
INSERT INTO stock_movements (product_id, type, qty, note, created_at) VALUES
(1,  'in',  40, 'Initial stock', now() - interval '20 days'),
(2,  'in',   8, 'Initial stock', now() - interval '20 days'),
(3,  'in',  60, 'Initial stock', now() - interval '20 days'),
(4,  'in',   5, 'Initial stock', now() - interval '20 days'),
(5,  'in',  75, 'Initial stock', now() - interval '20 days'),
(6,  'in',  25, 'Initial stock', now() - interval '20 days'),
(8,  'in',  30, 'Initial stock', now() - interval '20 days'),
(9,  'in', 120, 'Initial stock', now() - interval '20 days'),
(10, 'in',   9, 'Initial stock', now() - interval '20 days'),
(11, 'in',  45, 'Initial stock', now() - interval '20 days'),
(13, 'in', 200, 'Initial stock', now() - interval '20 days'),
(14, 'in', 150, 'Initial stock', now() - interval '20 days');

-- Sample orders (both still CREATED).
INSERT INTO sales_orders (customer, status, total, created_at) VALUES
('Toko Maju Jaya',   'CREATED', 195000, now() - interval '2 days'),
('Warung Sejahtera', 'CREATED', 106000, now() - interval '1 day');

INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES
(1, 'SKU-0001',  2, 75000),
(1, 'SKU-0006',  1, 45000),
(2, 'SKU-0003',  4, 16500),
(2, 'SKU-0013', 10,  4000);

-- The "out" movements that correspond to the two orders.
INSERT INTO stock_movements (product_id, type, qty, note, created_at) VALUES
(1,  'out',  2, 'Order #1', now() - interval '2 days'),
(6,  'out',  1, 'Order #1', now() - interval '2 days'),
(3,  'out',  4, 'Order #2', now() - interval '1 day'),
(13, 'out', 10, 'Order #2', now() - interval '1 day');
