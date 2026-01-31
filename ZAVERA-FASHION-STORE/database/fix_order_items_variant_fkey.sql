-- Fix order_items foreign key constraint
-- The constraint was pointing to the wrong table (variants instead of product_variants)

-- Drop the incorrect foreign key constraint
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_variant_id_fkey;

-- Add the correct foreign key constraint pointing to product_variants
ALTER TABLE order_items 
ADD CONSTRAINT order_items_variant_id_fkey 
FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE SET NULL;

-- Verify the fix
SELECT 
    tc.constraint_name, 
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM information_schema.table_constraints AS tc 
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
    AND tc.table_name='order_items'
    AND kcu.column_name='variant_id';
