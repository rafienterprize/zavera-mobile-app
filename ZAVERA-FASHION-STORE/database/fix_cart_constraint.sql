-- Fix Cart Constraint untuk Support Varian
-- Masalah: Constraint lama tidak mengizinkan produk sama dengan varian berbeda

-- 1. Hapus constraint lama
ALTER TABLE cart_items 
DROP CONSTRAINT IF EXISTS cart_items_cart_id_product_id_key;

-- 2. Hapus duplicate items (jika ada)
DELETE FROM cart_items a USING cart_items b
WHERE a.id < b.id 
AND a.cart_id = b.cart_id 
AND a.product_id = b.product_id
AND a.metadata = b.metadata;

-- 3. Tambah constraint baru yang include metadata (optional, tapi recommended)
-- Note: PostgreSQL tidak support unique constraint pada JSONB secara langsung
-- Jadi kita biarkan tanpa constraint, validasi di aplikasi level

-- 4. Verify
SELECT 
    cart_id,
    product_id,
    metadata->>'selected_size' as size,
    quantity,
    created_at
FROM cart_items
ORDER BY cart_id, product_id, created_at;

-- Expected: Bisa ada multiple rows dengan cart_id dan product_id sama
-- tapi metadata berbeda (size berbeda)
