-- Test adding M variant to cart 2 (which already has XL and L)
INSERT INTO cart_items (cart_id, product_id, quantity, price_snapshot, metadata) 
VALUES (2, 47, 3, 250000, '{"selected_size":"M","selected_color":"Black"}'::jsonb);

-- Check results
SELECT 
    ci.id, 
    ci.cart_id, 
    ci.product_id, 
    ci.quantity, 
    ci.metadata->>'selected_size' as size,
    ci.metadata->>'selected_color' as color
FROM cart_items ci 
WHERE ci.cart_id = 2 
ORDER BY ci.id;
