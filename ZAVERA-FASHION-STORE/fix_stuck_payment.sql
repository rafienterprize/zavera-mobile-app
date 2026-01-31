-- Fix stuck payment for order ZVR-20260122-26A6A050
-- Midtrans shows PAID (Settlement) but database shows PENDING

BEGIN;

-- Update payment status to PAID
UPDATE order_payments 
SET payment_status = 'PAID', 
    paid_at = NOW(), 
    updated_at = NOW()
WHERE order_id = 32 AND payment_status = 'PENDING';

-- Update order status to PAID
UPDATE orders 
SET status = 'PAID', 
    paid_at = NOW(), 
    updated_at = NOW()
WHERE id = 32 AND status = 'PENDING';

-- Verify the update
SELECT 
    o.id, 
    o.order_code, 
    o.status as order_status, 
    o.paid_at as order_paid_at,
    op.payment_status, 
    op.paid_at as payment_paid_at
FROM orders o
LEFT JOIN order_payments op ON o.id = op.order_id
WHERE o.id = 32;

COMMIT;
