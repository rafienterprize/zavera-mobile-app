-- Create Sample Audit Logs for Testing
-- Run this to populate audit logs page

-- Insert sample audit logs
INSERT INTO admin_audit_logs (
    admin_id,
    admin_email,
    action_type,
    action_detail,
    target_type,
    target_id,
    success,
    created_at
)
SELECT 
    1, -- Assuming admin ID 1 exists
    'admin@zavera.com',
    action_types.type,
    action_types.detail,
    'ORDER',
    (SELECT id FROM orders ORDER BY random() LIMIT 1),
    true,
    NOW() - (random() * INTERVAL '30 days')
FROM (
    VALUES 
        ('FORCE_CANCEL', 'Cancelled order due to payment timeout'),
        ('FORCE_REFUND', 'Processed refund for cancelled order'),
        ('RECONCILE_PAYMENT', 'Manually verified payment status'),
        ('SYNC_PAYMENT', 'Synchronized payment with gateway'),
        ('FORCE_CANCEL', 'Cancelled duplicate order'),
        ('RECONCILE_PAYMENT', 'Updated payment status after verification'),
        ('FORCE_REFUND', 'Issued refund for damaged product'),
        ('SYNC_PAYMENT', 'Fixed payment sync issue'),
        ('FORCE_CANCEL', 'Cancelled fraudulent order'),
        ('RECONCILE_PAYMENT', 'Resolved payment discrepancy')
) AS action_types(type, detail)
ON CONFLICT DO NOTHING;

-- Add some shipment-related audit logs
INSERT INTO admin_audit_logs (
    admin_id,
    admin_email,
    action_type,
    action_detail,
    target_type,
    target_id,
    success,
    created_at
)
SELECT 
    1,
    'admin@zavera.com',
    'FORCE_RESHIP',
    'Created replacement shipment for lost package',
    'SHIPMENT',
    (SELECT id FROM shipments ORDER BY random() LIMIT 1),
    true,
    NOW() - (random() * INTERVAL '20 days')
FROM generate_series(1, 5)
ON CONFLICT DO NOTHING;

-- Add some dispute resolution logs
INSERT INTO admin_audit_logs (
    admin_id,
    admin_email,
    action_type,
    action_detail,
    target_type,
    target_id,
    success,
    created_at
)
VALUES 
    (1, 'admin@zavera.com', 'RESOLVE_DISPUTE', 'Resolved dispute with refund', 'DISPUTE', 1, true, NOW() - INTERVAL '5 days'),
    (1, 'admin@zavera.com', 'RESOLVE_DISPUTE', 'Resolved dispute with replacement', 'DISPUTE', 2, true, NOW() - INTERVAL '3 days')
ON CONFLICT DO NOTHING;

SELECT 'Sample audit logs created successfully!' as message;
SELECT COUNT(*) as total_audit_logs FROM admin_audit_logs;
SELECT action_type, COUNT(*) as count 
FROM admin_audit_logs 
GROUP BY action_type 
ORDER BY count DESC;
