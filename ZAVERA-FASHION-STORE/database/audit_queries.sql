-- ============================================
-- ZAVERA DATABASE AUDIT QUERIES
-- Run these queries to check data integrity
-- ============================================

-- ============================================
-- 1. ORPHAN DETECTION
-- ============================================

-- 1.1 Orders without Payments (should be empty for paid orders)
SELECT 
    'ORPHAN_ORDER_NO_PAYMENT' as issue_type,
    o.id, 
    o.order_code, 
    o.status, 
    o.total_amount, 
    o.created_at
FROM orders o
LEFT JOIN payments p ON o.id = p.order_id
WHERE p.id IS NULL
AND o.status NOT IN ('CANCELLED', 'FAILED', 'EXPIRED')
AND o.created_at > NOW() - INTERVAL '30 days';

-- 1.2 Payments without Orders (should be empty)
SELECT 
    'ORPHAN_PAYMENT_NO_ORDER' as issue_type,
    p.id, 
    p.external_id, 
    p.amount, 
    p.status, 
    p.created_at
FROM payments p
LEFT JOIN orders o ON p.order_id = o.id
WHERE o.id IS NULL;

-- 1.3 Shipments without Orders (should be empty)
SELECT 
    'ORPHAN_SHIPMENT_NO_ORDER' as issue_type,
    s.id,
    s.order_id,
    s.tracking_number,
    s.status,
    s.created_at
FROM shipments s
LEFT JOIN orders o ON s.order_id = o.id
WHERE o.id IS NULL;

-- 1.4 Refunds without Payments (should be empty)
SELECT 
    'ORPHAN_REFUND_NO_PAYMENT' as issue_type,
    r.id,
    r.refund_code,
    r.order_id,
    r.payment_id,
    r.refund_amount,
    r.status
FROM refunds r
LEFT JOIN payments p ON r.payment_id = p.id
WHERE r.payment_id IS NOT NULL AND p.id IS NULL;

-- 1.5 Disputes without Shipments (for shipment-related disputes)
SELECT 
    'ORPHAN_DISPUTE_NO_SHIPMENT' as issue_type,
    d.id,
    d.dispute_code,
    d.shipment_id,
    d.status
FROM disputes d
LEFT JOIN shipments s ON d.shipment_id = s.id
WHERE d.shipment_id IS NOT NULL AND s.id IS NULL;

-- ============================================
-- 2. DATA INTEGRITY CHECKS
-- ============================================

-- 2.1 Over-Refunded Orders
SELECT 
    'OVER_REFUND' as issue_type,
    o.id, 
    o.order_code, 
    o.total_amount,
    p.amount as paid_amount,
    COALESCE(SUM(r.refund_amount), 0) as total_refunded,
    COALESCE(SUM(r.refund_amount), 0) - p.amount as over_refund_amount
FROM orders o
JOIN payments p ON o.id = p.order_id AND p.status = 'SUCCESS'
LEFT JOIN refunds r ON o.id = r.order_id AND r.status = 'COMPLETED'
GROUP BY o.id, o.order_code, o.total_amount, p.amount
HAVING COALESCE(SUM(r.refund_amount), 0) > p.amount;

-- 2.2 Negative Stock Products
SELECT 
    'NEGATIVE_STOCK' as issue_type,
    id, 
    name, 
    slug, 
    stock 
FROM products 
WHERE stock < 0;

-- 2.3 Duplicate Gateway Transaction IDs
SELECT 
    'DUPLICATE_TRANSACTION_ID' as issue_type,
    transaction_id, 
    COUNT(*) as count,
    array_agg(id) as payment_ids
FROM payments
WHERE transaction_id IS NOT NULL AND transaction_id != ''
GROUP BY transaction_id
HAVING COUNT(*) > 1;

-- 2.4 Status Mismatches
SELECT 
    'STATUS_MISMATCH' as issue_type,
    o.id,
    o.order_code,
    o.status as order_status,
    p.status as payment_status,
    s.status as shipment_status,
    CASE 
        WHEN o.status = 'PAID' AND p.status != 'SUCCESS' THEN 'Order PAID but payment not SUCCESS'
        WHEN o.status = 'SHIPPED' AND s.status NOT IN ('SHIPPED', 'IN_TRANSIT', 'OUT_FOR_DELIVERY', 'DELIVERED') THEN 'Order SHIPPED but shipment status mismatch'
        WHEN p.status = 'SUCCESS' AND o.status = 'PENDING' THEN 'Payment SUCCESS but order still PENDING'
        ELSE 'Unknown mismatch'
    END as mismatch_reason
FROM orders o
LEFT JOIN payments p ON o.id = p.order_id
LEFT JOIN shipments s ON o.id = s.order_id
WHERE 
    (o.status = 'PAID' AND p.status != 'SUCCESS')
    OR (o.status = 'SHIPPED' AND s.status NOT IN ('SHIPPED', 'IN_TRANSIT', 'OUT_FOR_DELIVERY', 'DELIVERED'))
    OR (p.status = 'SUCCESS' AND o.status = 'PENDING');

-- ============================================
-- 3. STUCK/PROBLEMATIC RECORDS
-- ============================================

-- 3.1 Stuck Shipments (No Update > 7 Days)
SELECT 
    'STUCK_SHIPMENT' as issue_type,
    s.id,
    s.order_id,
    o.order_code,
    s.tracking_number,
    s.status,
    s.provider_code,
    s.updated_at,
    EXTRACT(DAY FROM (NOW() - s.updated_at)) as days_stuck
FROM shipments s
JOIN orders o ON s.order_id = o.id
WHERE s.status IN ('SHIPPED', 'IN_TRANSIT', 'OUT_FOR_DELIVERY')
AND s.updated_at < NOW() - INTERVAL '7 days'
ORDER BY s.updated_at ASC;

-- 3.2 Pending Payments > 2 Hours
SELECT 
    'STUCK_PAYMENT' as issue_type,
    p.id,
    p.order_id,
    o.order_code,
    p.amount,
    p.status,
    p.created_at,
    EXTRACT(HOUR FROM (NOW() - p.created_at)) as hours_pending
FROM payments p
JOIN orders o ON p.order_id = o.id
WHERE p.status = 'PENDING'
AND p.created_at < NOW() - INTERVAL '2 hours'
ORDER BY p.created_at ASC;

-- 3.3 Reship Loops (More than 3 Reships)
SELECT 
    'RESHIP_LOOP' as issue_type,
    original_shipment_id,
    COUNT(*) as reship_count,
    array_agg(id) as shipment_chain
FROM shipments
WHERE is_replacement = true
GROUP BY original_shipment_id
HAVING COUNT(*) > 3;

-- 3.4 Unresolved Disputes > 7 Days
SELECT 
    'STALE_DISPUTE' as issue_type,
    d.id,
    d.dispute_code,
    d.order_id,
    o.order_code,
    d.status,
    d.created_at,
    EXTRACT(DAY FROM (NOW() - d.created_at)) as days_open
FROM disputes d
JOIN orders o ON d.order_id = o.id
WHERE d.status IN ('OPEN', 'INVESTIGATING', 'EVIDENCE_REQUIRED', 'PENDING_RESOLUTION')
AND d.created_at < NOW() - INTERVAL '7 days'
ORDER BY d.created_at ASC;

-- ============================================
-- 4. FINANCIAL SUMMARY
-- ============================================

-- 4.1 Daily Revenue Summary (Last 30 Days)
SELECT 
    DATE(o.paid_at) as date,
    COUNT(*) as order_count,
    SUM(o.total_amount) as total_revenue,
    AVG(o.total_amount) as avg_order_value
FROM orders o
WHERE o.status IN ('PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
AND o.paid_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(o.paid_at)
ORDER BY date DESC;

-- 4.2 Refund Summary (Last 30 Days)
SELECT 
    DATE(r.completed_at) as date,
    COUNT(*) as refund_count,
    SUM(r.refund_amount) as total_refunded,
    r.reason
FROM refunds r
WHERE r.status = 'COMPLETED'
AND r.completed_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(r.completed_at), r.reason
ORDER BY date DESC;

-- 4.3 Net Revenue (Revenue - Refunds)
SELECT 
    COALESCE(SUM(CASE WHEN o.status IN ('PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED') THEN o.total_amount ELSE 0 END), 0) as gross_revenue,
    COALESCE(SUM(r.refund_amount), 0) as total_refunds,
    COALESCE(SUM(CASE WHEN o.status IN ('PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED') THEN o.total_amount ELSE 0 END), 0) - COALESCE(SUM(r.refund_amount), 0) as net_revenue
FROM orders o
LEFT JOIN refunds r ON o.id = r.order_id AND r.status = 'COMPLETED'
WHERE o.created_at > NOW() - INTERVAL '30 days';

-- ============================================
-- 5. SYSTEM HEALTH METRICS
-- ============================================

-- 5.1 Order Status Distribution
SELECT 
    status,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM orders
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY status
ORDER BY count DESC;

-- 5.2 Payment Status Distribution
SELECT 
    status,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM payments
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY status
ORDER BY count DESC;

-- 5.3 Shipment Status Distribution
SELECT 
    status,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM shipments
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY status
ORDER BY count DESC;

-- 5.4 Average Delivery Time
SELECT 
    provider_code,
    COUNT(*) as delivered_count,
    ROUND(AVG(EXTRACT(DAY FROM (delivered_at - shipped_at))), 2) as avg_delivery_days
FROM shipments
WHERE status = 'DELIVERED' 
AND delivered_at IS NOT NULL 
AND shipped_at IS NOT NULL
AND delivered_at > NOW() - INTERVAL '30 days'
GROUP BY provider_code
ORDER BY avg_delivery_days ASC;

-- ============================================
-- END OF AUDIT QUERIES
-- ============================================
SELECT 'Audit queries completed. Review results above for any issues.' as status;
