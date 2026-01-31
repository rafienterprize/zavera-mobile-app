@echo off
echo Adding more sample shipments...
echo.

psql -U postgres -d zavera_db -c "INSERT INTO shipments (order_id, provider_code, provider_name, service_code, service_name, tracking_number, status, cost, weight, origin_city_name, destination_city_name, days_without_update, created_at, updated_at, shipped_at, delivered_at) SELECT o.id, CASE WHEN random() < 0.2 THEN 'jne' WHEN random() < 0.4 THEN 'jnt' WHEN random() < 0.6 THEN 'sicepat' WHEN random() < 0.8 THEN 'anteraja' ELSE 'tiki' END, CASE WHEN random() < 0.2 THEN 'JNE' WHEN random() < 0.4 THEN 'J&T Express' WHEN random() < 0.6 THEN 'SiCepat' WHEN random() < 0.8 THEN 'AnterAja' ELSE 'TIKI' END, 'REG', 'Regular', 'TRACK' || LPAD((random() * 10000000)::INT::TEXT, 10, '0'), CASE WHEN random() < 0.5 THEN 'DELIVERED'::shipment_status WHEN random() < 0.7 THEN 'SHIPPED'::shipment_status WHEN random() < 0.85 THEN 'IN_TRANSIT'::shipment_status ELSE 'OUT_FOR_DELIVERY'::shipment_status END, 10000 + (random() * 30000)::INT, 500 + (random() * 3000)::INT, 'Jakarta', 'Bandung', (random() * 5)::INT, NOW() - (random() * INTERVAL '30 days'), NOW() - (random() * INTERVAL '10 days'), NOW() - (random() * INTERVAL '25 days'), CASE WHEN random() < 0.5 THEN NOW() - (random() * INTERVAL '5 days') ELSE NULL END FROM orders o WHERE o.status IN ('PAID'::order_status, 'SHIPPED'::order_status, 'DELIVERED'::order_status) AND NOT EXISTS (SELECT 1 FROM shipments s WHERE s.order_id = o.id) LIMIT 30;"

echo.
echo Done! Checking shipments count...
psql -U postgres -d zavera_db -c "SELECT COUNT(*) as total_shipments FROM shipments;"
psql -U postgres -d zavera_db -c "SELECT status, COUNT(*) as count FROM shipments GROUP BY status ORDER BY count DESC;"

pause
