-- ============================================
-- ZAVERA COMMERCE PLATFORM UPGRADE MIGRATION
-- Production-grade e-commerce system upgrade
-- ============================================
-- This migration adds:
-- 1. PACKING and REFUNDED order statuses
-- 2. Resi and delivery tracking columns
-- 3. Stock movements table
-- 4. Shipping snapshots table
-- 5. Email templates and logs tables
-- ============================================

-- ============================================
-- 1. EXTEND ORDER STATUS ENUM
-- Add PACKING (between PAID and SHIPPED)
-- Add REFUNDED (terminal state)
-- ============================================

-- Check if PACKING exists, if not add it
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'PACKING' AND enumtypid = 'order_status'::regtype) THEN
        ALTER TYPE order_status ADD VALUE 'PACKING' AFTER 'PAID';
    END IF;
END $$;

-- Check if REFUNDED exists, if not add it
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'REFUNDED' AND enumtypid = 'order_status'::regtype) THEN
        ALTER TYPE order_status ADD VALUE 'REFUNDED' AFTER 'CANCELLED';
    END IF;
END $$;

-- Check if EXPIRED exists, if not add it
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'EXPIRED' AND enumtypid = 'order_status'::regtype) THEN
        ALTER TYPE order_status ADD VALUE 'EXPIRED' AFTER 'FAILED';
    END IF;
END $$;

-- Check if COMPLETED exists, if not add it
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'COMPLETED' AND enumtypid = 'order_status'::regtype) THEN
        ALTER TYPE order_status ADD VALUE 'COMPLETED' AFTER 'DELIVERED';
    END IF;
END $$;

-- ============================================
-- 2. ADD MISSING COLUMNS TO ORDERS TABLE
-- ============================================

-- Add resi column for airway bill number
ALTER TABLE orders ADD COLUMN IF NOT EXISTS resi VARCHAR(100);
CREATE INDEX IF NOT EXISTS idx_orders_resi ON orders(resi) WHERE resi IS NOT NULL;

-- Add delivered_at timestamp
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP;

-- Add origin and destination city columns
ALTER TABLE orders ADD COLUMN IF NOT EXISTS origin_city VARCHAR(100) DEFAULT 'Semarang';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS destination_city VARCHAR(100);

-- Add stock_reserved column if not exists
ALTER TABLE orders ADD COLUMN IF NOT EXISTS stock_reserved BOOLEAN DEFAULT true;

-- ============================================
-- 3. CREATE STOCK MOVEMENTS TABLE
-- Tracks all stock operations for audit
-- ============================================

CREATE TABLE IF NOT EXISTS stock_movements (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    movement_type VARCHAR(20) NOT NULL,
    quantity INTEGER NOT NULL,
    balance_after INTEGER NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_movement_type CHECK (movement_type IN ('RESERVE', 'RELEASE', 'DEDUCT', 'ADJUSTMENT'))
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_product ON stock_movements(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_order ON stock_movements(order_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_type ON stock_movements(movement_type);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created ON stock_movements(created_at DESC);

COMMENT ON TABLE stock_movements IS 'Audit trail for all stock operations';
COMMENT ON COLUMN stock_movements.movement_type IS 'RESERVE=checkout, RELEASE=cancel/expire, DEDUCT=permanent after payment';
COMMENT ON COLUMN stock_movements.balance_after IS 'Product stock balance after this movement';

-- ============================================
-- 4. CREATE SHIPPING SNAPSHOTS TABLE
-- Stores RajaOngkir response at checkout time
-- ============================================

CREATE TABLE IF NOT EXISTS shipping_snapshots (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    courier VARCHAR(50) NOT NULL,
    service VARCHAR(100) NOT NULL,
    cost DECIMAL(12, 2) NOT NULL,
    etd VARCHAR(50) NOT NULL,
    origin_city_id VARCHAR(20) NOT NULL,
    origin_city_name VARCHAR(100),
    destination_city_id VARCHAR(20) NOT NULL,
    destination_city_name VARCHAR(100),
    destination_district_id VARCHAR(20),
    weight INTEGER NOT NULL,
    rajaongkir_raw_json JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_shipping_snapshot_order UNIQUE (order_id),
    CONSTRAINT chk_shipping_cost CHECK (cost >= 0),
    CONSTRAINT chk_shipping_weight CHECK (weight > 0)
);

CREATE INDEX IF NOT EXISTS idx_shipping_snapshots_order ON shipping_snapshots(order_id);
CREATE INDEX IF NOT EXISTS idx_shipping_snapshots_courier ON shipping_snapshots(courier);

COMMENT ON TABLE shipping_snapshots IS 'Immutable record of shipping cost from RajaOngkir at checkout';
COMMENT ON COLUMN shipping_snapshots.rajaongkir_raw_json IS 'Complete API response for audit';

-- ============================================
-- 5. CREATE EMAIL TEMPLATES TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS email_templates (
    id SERIAL PRIMARY KEY,
    template_key VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    subject_template TEXT NOT NULL,
    html_template TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_email_templates_key ON email_templates(template_key);

-- Trigger for updated_at
DROP TRIGGER IF EXISTS update_email_templates_updated_at ON email_templates;
CREATE TRIGGER update_email_templates_updated_at 
    BEFORE UPDATE ON email_templates 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE email_templates IS 'HTML email templates for transactional emails';

-- ============================================
-- 6. CREATE EMAIL LOGS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS email_logs (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    template_key VARCHAR(50) NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_email_status CHECK (status IN ('PENDING', 'SENT', 'FAILED', 'RETRY'))
);

CREATE INDEX IF NOT EXISTS idx_email_logs_order ON email_logs(order_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_user ON email_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status);
CREATE INDEX IF NOT EXISTS idx_email_logs_template ON email_logs(template_key);
CREATE INDEX IF NOT EXISTS idx_email_logs_created ON email_logs(created_at DESC);

COMMENT ON TABLE email_logs IS 'Log of all transactional emails sent';

-- ============================================
-- 7. SEED EMAIL TEMPLATES
-- ============================================

INSERT INTO email_templates (template_key, name, subject_template, html_template, is_active) VALUES
(
    'ORDER_CREATED',
    'Order Created',
    '🛍️ Pesanan ZAVERA #{{.OrderCode}} telah dibuat',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; }
        .header { background: #000; color: #fff; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .order-info { background: #f9f9f9; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .items-table { width: 100%; border-collapse: collapse; margin: 15px 0; }
        .items-table th, .items-table td { padding: 10px; border-bottom: 1px solid #eee; text-align: left; }
        .total-row { font-weight: bold; background: #f0f0f0; }
        .footer { background: #f5f5f5; padding: 15px; text-align: center; font-size: 12px; color: #666; }
        .btn { display: inline-block; background: #000; color: #fff; padding: 12px 24px; text-decoration: none; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>ZAVERA</h1>
    </div>
    <div class="content">
        <h2>Pesanan Anda Telah Dibuat!</h2>
        <p>Halo {{.CustomerName}},</p>
        <p>Terima kasih telah berbelanja di ZAVERA. Pesanan Anda telah berhasil dibuat.</p>
        
        <div class="order-info">
            <strong>Nomor Pesanan:</strong> {{.OrderCode}}<br>
            <strong>Tanggal:</strong> {{.CreatedAt}}<br>
            <strong>Status:</strong> Menunggu Pembayaran
        </div>
        
        <h3>Detail Pesanan</h3>
        <table class="items-table">
            <tr><th>Produk</th><th>Qty</th><th>Harga</th></tr>
            {{range .Items}}
            <tr><td>{{.ProductName}}</td><td>{{.Quantity}}</td><td>Rp {{.Subtotal}}</td></tr>
            {{end}}
            <tr><td colspan="2">Subtotal</td><td>Rp {{.Subtotal}}</td></tr>
            <tr><td colspan="2">Ongkir ({{.Courier}} - {{.Service}})</td><td>Rp {{.ShippingCost}}</td></tr>
            <tr class="total-row"><td colspan="2">Total</td><td>Rp {{.TotalAmount}}</td></tr>
        </table>
        
        <h3>Alamat Pengiriman</h3>
        <p>{{.ShippingAddress}}</p>
        
        <h3>Instruksi Pembayaran</h3>
        <p>Silakan selesaikan pembayaran dalam waktu 24 jam untuk memproses pesanan Anda.</p>
        <p style="text-align: center; margin: 20px 0;">
            <a href="{{.PaymentURL}}" class="btn">Bayar Sekarang</a>
        </p>
    </div>
    <div class="footer">
        <p>© 2026 ZAVERA. All rights reserved.</p>
        <p>Jika ada pertanyaan, hubungi kami di support@zavera.com</p>
    </div>
</body>
</html>',
    true
),
(
    'PAYMENT_SUCCESS',
    'Payment Success',
    '💳 Pembayaran diterima – Pesanan #{{.OrderCode}}',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; }
        .header { background: #000; color: #fff; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .success-badge { background: #4CAF50; color: #fff; padding: 15px; border-radius: 5px; text-align: center; margin: 15px 0; }
        .order-info { background: #f9f9f9; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .footer { background: #f5f5f5; padding: 15px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>ZAVERA</h1>
    </div>
    <div class="content">
        <div class="success-badge">
            <h2>✓ Pembayaran Berhasil!</h2>
        </div>
        <p>Halo {{.CustomerName}},</p>
        <p>Pembayaran untuk pesanan Anda telah kami terima. Pesanan Anda sedang diproses dan akan segera dikirim.</p>
        
        <div class="order-info">
            <strong>Nomor Pesanan:</strong> {{.OrderCode}}<br>
            <strong>Total Pembayaran:</strong> Rp {{.TotalAmount}}<br>
            <strong>Metode Pembayaran:</strong> {{.PaymentMethod}}<br>
            <strong>Waktu Pembayaran:</strong> {{.PaidAt}}
        </div>
        
        <p>Kami akan mengirimkan email konfirmasi pengiriman setelah pesanan Anda dikirim.</p>
    </div>
    <div class="footer">
        <p>© 2026 ZAVERA. All rights reserved.</p>
    </div>
</body>
</html>',
    true
),
(
    'ORDER_SHIPPED',
    'Order Shipped',
    '📦 Pesanan #{{.OrderCode}} sedang dikirim',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; }
        .header { background: #000; color: #fff; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .shipping-badge { background: #2196F3; color: #fff; padding: 15px; border-radius: 5px; text-align: center; margin: 15px 0; }
        .tracking-box { background: #f9f9f9; padding: 20px; border-radius: 5px; margin: 15px 0; text-align: center; }
        .tracking-number { font-size: 24px; font-weight: bold; color: #000; letter-spacing: 2px; }
        .btn { display: inline-block; background: #000; color: #fff; padding: 12px 24px; text-decoration: none; border-radius: 5px; }
        .footer { background: #f5f5f5; padding: 15px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>ZAVERA</h1>
    </div>
    <div class="content">
        <div class="shipping-badge">
            <h2>📦 Pesanan Anda Sedang Dikirim!</h2>
        </div>
        <p>Halo {{.CustomerName}},</p>
        <p>Kabar baik! Pesanan Anda telah dikirim dan sedang dalam perjalanan.</p>
        
        <div class="tracking-box">
            <p><strong>Kurir:</strong> {{.Courier}} - {{.Service}}</p>
            <p><strong>Nomor Resi:</strong></p>
            <p class="tracking-number">{{.Resi}}</p>
            <p><strong>Estimasi Tiba:</strong> {{.ETD}}</p>
        </div>
        
        <p style="text-align: center; margin: 20px 0;">
            <a href="{{.TrackingURL}}" class="btn">Lacak Pengiriman</a>
        </p>
        
        <p><strong>Alamat Pengiriman:</strong><br>{{.ShippingAddress}}</p>
    </div>
    <div class="footer">
        <p>© 2026 ZAVERA. All rights reserved.</p>
    </div>
</body>
</html>',
    true
),
(
    'ORDER_DELIVERED',
    'Order Delivered',
    '🎉 Pesanan #{{.OrderCode}} sudah sampai',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; }
        .header { background: #000; color: #fff; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .delivered-badge { background: #4CAF50; color: #fff; padding: 20px; border-radius: 5px; text-align: center; margin: 15px 0; }
        .order-info { background: #f9f9f9; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .btn { display: inline-block; background: #000; color: #fff; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 5px; }
        .footer { background: #f5f5f5; padding: 15px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>ZAVERA</h1>
    </div>
    <div class="content">
        <div class="delivered-badge">
            <h2>🎉 Pesanan Telah Sampai!</h2>
        </div>
        <p>Halo {{.CustomerName}},</p>
        <p>Pesanan Anda telah berhasil diterima. Terima kasih telah berbelanja di ZAVERA!</p>
        
        <div class="order-info">
            <strong>Nomor Pesanan:</strong> {{.OrderCode}}<br>
            <strong>Tanggal Diterima:</strong> {{.DeliveredAt}}<br>
            <strong>Kurir:</strong> {{.Courier}} - {{.Service}}
        </div>
        
        <p>Kami harap Anda puas dengan produk yang Anda terima. Jika ada masalah dengan pesanan, silakan hubungi kami dalam 7 hari.</p>
        
        <p style="text-align: center; margin: 20px 0;">
            <a href="{{.ReviewURL}}" class="btn">Beri Ulasan</a>
            <a href="{{.ShopURL}}" class="btn">Belanja Lagi</a>
        </p>
    </div>
    <div class="footer">
        <p>© 2026 ZAVERA. All rights reserved.</p>
    </div>
</body>
</html>',
    true
)
ON CONFLICT (template_key) DO UPDATE SET
    name = EXCLUDED.name,
    subject_template = EXCLUDED.subject_template,
    html_template = EXCLUDED.html_template,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================
-- 8. ADD UNIQUE CONSTRAINT ON RESI
-- ============================================

-- Create unique index on resi (only for non-null values)
DROP INDEX IF EXISTS idx_orders_resi_unique;
CREATE UNIQUE INDEX idx_orders_resi_unique ON orders(resi) WHERE resi IS NOT NULL;

-- ============================================
-- VERIFICATION
-- ============================================

SELECT 'ZAVERA Commerce Upgrade Migration completed successfully' AS status;
SELECT enumlabel FROM pg_enum WHERE enumtypid = 'order_status'::regtype ORDER BY enumsortorder;
