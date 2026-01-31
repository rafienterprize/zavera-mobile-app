-- Create Sample Shipments for Testing - FIXED VERSION
-- Run this to populate shipments page with realistic data

-- First, create sample orders if they don't exist
DO $$
DECLARE
    i INT;
    order_code TEXT;
    order_status order_status;
    statuses order_status[] := ARRAY['PAID'::order_status, 'PACKING'::order_status, 'SHIPPED'::order_status, 'DELIVERED'::order_status];
BEGIN
    FOR i IN 1..50 LOOP
        order_code := 'ORD-SAMPLE-' || LPAD(i::TEXT, 4, '0');
        order_status := statuses[1 + floor(random() * 4)];
        
        INSERT INTO orders (
            order_code, 
            customer_name, 
            customer_email, 
            customer_phone, 
            subtotal, 
            shipping_cost, 
            total_amount, 
            status, 
            created_at, 
            updated_at,
            paid_at
        )
        VALUES (
            order_code,
            'Sample Customer ' || i,
            'customer' || i || '@example.com',
            '0812345678' || LPAD(i::TEXT, 2, '0'),
            300000 + (random() * 700000)::INT,
            10000 + (random() * 20000)::INT,
            310000 + (random() * 720000)::INT,
            order_status,
            NOW() - (random() * INTERVAL '60 days'),
            NOW() - (random() * INTERVAL '30 days'),
            CASE WHEN order_status != 'PENDING'::order_status THEN NOW() - (random() * INTERVAL '50 days') ELSE NULL END
        )
        ON CONFLICT (order_code) DO NOTHING;
    END LOOP;
END $$;

-- Create shipments for orders
INSERT INTO shipments (
    order_id, 
    provider_code, 
    provider_name, 
    service_code, 
    service_name, 
    tracking_number, 
    status, 
    cost, 
    weight, 
    origin_city_name, 
    destination_city_name,
    days_without_update,
    created_at, 
    updated_at,
    shipped_at,
    delivered_at
)
SELECT 
    o.id,
    CASE 
        WHEN random() < 0.2 THEN 'jne'
        WHEN random() < 0.4 THEN 'jnt'
        WHEN random() < 0.6 THEN 'sicepat'
        WHEN random() < 0.8 THEN 'anteraja'
        ELSE 'tiki'
    END,
    CASE 
        WHEN random() < 0.2 THEN 'JNE'
        WHEN random() < 0.4 THEN 'J&T Express'
        WHEN random() < 0.6 THEN 'SiCepat'
        WHEN random() < 0.8 THEN 'AnterAja'
        ELSE 'TIKI'
    END,
    CASE 
        WHEN random() < 0.5 THEN 'REG'
        ELSE 'YES'
    END,
    CASE 
        WHEN random() < 0.5 THEN 'Regular'
        ELSE 'Express'
    END,
    CASE 
        WHEN random() < 0.2 THEN 'JNE' || LPAD((random() * 10000000)::INT::TEXT, 10, '0')
        WHEN random() < 0.4 THEN 'JT' || LPAD((random() * 10000000)::INT::TEXT, 10, '0')
        WHEN random() < 0.6 THEN 'SC' || LPAD((random() * 10000000)::INT::TEXT, 10, '0')
        WHEN random() < 0.8 THEN 'AA' || LPAD((random() * 10000000)::INT::TEXT, 10, '0')
        ELSE 'TK' || LPAD((random() * 10000000)::INT::TEXT, 10, '0')
    END,
    CASE 
        WHEN random() < 0.6 THEN 'DELIVERED'::shipment_status
        WHEN random() < 0.75 THEN 'SHIPPED'::shipment_status
        WHEN random() < 0.85 THEN 'IN_TRANSIT'::shipment_status
        WHEN random() < 0.92 THEN 'OUT_FOR_DELIVERY'::shipment_status
        WHEN random() < 0.96 THEN 'PICKUP_SCHEDULED'::shipment_status
        ELSE 'DELIVERY_FAILED'::shipment_status
    END,
    10000 + (random() * 30000)::INT,
    500 + (random() * 3000)::INT,
    CASE 
        WHEN random() < 0.5 THEN 'Jakarta'
        ELSE 'Surabaya'
    END,
    CASE 
        WHEN random() < 0.2 THEN 'Bandung'
        WHEN random() < 0.4 THEN 'Semarang'
        WHEN random() < 0.6 THEN 'Yogyakarta'
        WHEN random() < 0.8 THEN 'Malang'
        ELSE 'Bali'
    END,
    (random() * 10)::INT,
    o.created_at + INTERVAL '1 day',
    NOW() - (random() * INTERVAL '20 days'),
    o.created_at + INTERVAL '2 days',
    CASE 
        WHEN random() < 0.6 THEN o.created_at + INTERVAL '5 days' + (random() * INTERVAL '3 days')
        ELSE NULL
    END
FROM orders o
WHERE o.order_code LIKE 'ORD-SAMPLE-%'
AND o.status IN ('PAID'::order_status, 'PACKING'::order_status, 'SHIPPED'::order_status, 'DELIVERED'::order_status)
AND NOT EXISTS (
    SELECT 1 FROM shipments s WHERE s.order_id = o.id
)
ON CONFLICT DO NOTHING;

-- Update some shipments to have realistic days_without_update
UPDATE shipments 
SET days_without_update = EXTRACT(DAY FROM (NOW() - updated_at))::INT
WHERE tracking_number LIKE 'JNE%' OR tracking_number LIKE 'JT%' OR tracking_number LIKE 'SC%';

-- Add some stuck shipments (7+ days no update)
WITH stuck_candidates AS (
    SELECT id FROM shipments 
    WHERE tracking_number LIKE 'AA%'
    AND random() < 0.3
    ORDER BY random()
    LIMIT 5
)
UPDATE shipments 
SET 
    days_without_update = 7 + (random() * 5)::INT,
    updated_at = NOW() - INTERVAL '7 days' - (random() * INTERVAL '5 days'),
    status = 'SHIPPED'::shipment_status
WHERE id IN (SELECT id FROM stuck_candidates);

-- Add some in-transit shipments
WITH transit_candidates AS (
    SELECT id FROM shipments 
    WHERE tracking_number LIKE 'TK%'
    AND random() < 0.5
    ORDER BY random()
    LIMIT 10
)
UPDATE shipments 
SET 
    status = 'IN_TRANSIT'::shipment_status,
    days_without_update = (random() * 3)::INT
WHERE id IN (SELECT id FROM transit_candidates);

SELECT 'Sample shipments created successfully!' as message;
SELECT COUNT(*) as total_shipments FROM shipments;
SELECT status, COUNT(*) as count FROM shipments GROUP BY status ORDER BY count DESC;
