-- Insert cart items dengan variant_id yang benar
-- L × 2 (variant_id = 5)
INSERT INTO cart_items (cart_id, product_id, variant_id, quantity, price_snapshot, metadata) 
VALUES (2, 47, 5, 2, 330000, '{"selected_size":"L","selected_color":"Black"}'::jsonb);

-- XL × 1 (variant_id = 6)
INSERT INTO cart_items (cart_id, product_id, variant_id, quantity, price_snapshot, metadata) 
VALUES (2, 47, 6, 1, 330000, '{"selected_size":"XL","selected_color":"Black"}'::jsonb);

-- Verify
SELECT 
    ci.id, 
    ci.cart_id, 
    ci.product_id,
    ci.variant_id,
    ci.quantity, 
    ci.metadata->>'selected_size' as size,
    v.stock_quantity as variant_stock
FROM cart_items ci 
LEFT JOIN product_variants v ON ci.variant_id = v.id
WHERE ci.cart_id = 2 
ORDER BY ci.id;
