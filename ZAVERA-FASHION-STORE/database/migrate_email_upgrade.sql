-- ============================================
-- ZAVERA EMAIL SYSTEM UPGRADE
-- Tokopedia-style transactional email policy
-- ============================================
-- Email hanya dikirim untuk event dengan dampak legal/finansial:
-- 1. ORDER_CREATED - Invoice dibuat
-- 2. PAYMENT_SUCCESS - Uang diterima
-- 3. ORDER_SHIPPED - Barang diserahkan ke kurir
-- 4. ORDER_DELIVERED - Barang diterima
-- 5. ORDER_CANCELLED - Kontrak dibatalkan
-- 6. ORDER_REFUNDED - Uang dikembalikan
-- ============================================

-- ============================================
-- 1. ADD UNIQUE CONSTRAINT TO PREVENT DUPLICATE EMAILS
-- ============================================
-- Prevent sending same email type for same order twice
CREATE UNIQUE INDEX IF NOT EXISTS idx_email_logs_unique_event 
ON email_logs(order_id, template_key) 
WHERE status = 'SENT';

-- ============================================
-- 2. ADD ORDER_CANCELLED EMAIL TEMPLATE
-- ============================================
INSERT INTO email_templates (template_key, name, subject_template, html_template, is_active) VALUES
(
    'ORDER_CANCELLED',
    'Order Cancelled',
    '❌ Pesanan #{{.OrderCode}} telah dibatalkan',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; }
        .header { background: #000; color: #fff; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .cancelled-badge { background: #EF4444; color: #fff; padding: 15px; border-radius: 5px; text-align: center; margin: 15px 0; }
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
        <div class="cancelled-badge">
            <h2>❌ Pesanan Dibatalkan</h2>
        </div>
        <p>Halo {{.CustomerName}},</p>
        <p>Pesanan Anda telah dibatalkan. {{.CancellationReason}}</p>
        
        <div class="order-info">
            <strong>Nomor Pesanan:</strong> {{.OrderCode}}<br>
            <strong>Tanggal Pesanan:</strong> {{.CreatedAt}}<br>
            <strong>Tanggal Dibatalkan:</strong> {{.CancelledAt}}<br>
            <strong>Alasan:</strong> {{.CancellationReason}}
        </div>
        
        <h3>Detail Pesanan yang Dibatalkan</h3>
        <table class="items-table">
            <tr><th>Produk</th><th>Qty</th><th>Harga</th></tr>
            {{range .Items}}
            <tr><td>{{.ProductName}}</td><td>{{.Quantity}}</td><td>Rp {{.Subtotal}}</td></tr>
            {{end}}
            <tr><td colspan="2">Subtotal</td><td>Rp {{.Subtotal}}</td></tr>
            <tr><td colspan="2">Ongkir</td><td>Rp {{.ShippingCost}}</td></tr>
            <tr class="total-row"><td colspan="2">Total</td><td>Rp {{.TotalAmount}}</td></tr>
        </table>
        
        <h3>Alamat Pengiriman</h3>
        <p>{{.ShippingAddress}}</p>
        
        {{if .RefundInfo}}
        <div class="order-info" style="background: #FEF3C7; border: 1px solid #F59E0B;">
            <strong>💰 Informasi Pengembalian Dana:</strong><br>
            {{.RefundInfo}}
        </div>
        {{end}}
        
        <p>Jika Anda memiliki pertanyaan, silakan hubungi customer service kami.</p>
        
        <p style="text-align: center; margin: 20px 0;">
            <a href="{{.ShopURL}}" class="btn">Belanja Lagi</a>
        </p>
    </div>
    <div class="footer">
        <p>© 2026 ZAVERA. All rights reserved.</p>
        <p>Jika ada pertanyaan, hubungi kami di support@zavera.com</p>
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
-- 3. ADD ORDER_REFUNDED EMAIL TEMPLATE
-- ============================================
INSERT INTO email_templates (template_key, name, subject_template, html_template, is_active) VALUES
(
    'ORDER_REFUNDED',
    'Order Refunded',
    '💰 Refund untuk Pesanan #{{.OrderCode}} telah diproses',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; }
        .header { background: #000; color: #fff; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .refund-badge { background: #10B981; color: #fff; padding: 15px; border-radius: 5px; text-align: center; margin: 15px 0; }
        .order-info { background: #f9f9f9; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .refund-box { background: #D1FAE5; border: 1px solid #10B981; padding: 20px; border-radius: 5px; margin: 15px 0; text-align: center; }
        .refund-amount { font-size: 28px; font-weight: bold; color: #059669; }
        .items-table { width: 100%; border-collapse: collapse; margin: 15px 0; }
        .items-table th, .items-table td { padding: 10px; border-bottom: 1px solid #eee; text-align: left; }
        .footer { background: #f5f5f5; padding: 15px; text-align: center; font-size: 12px; color: #666; }
        .btn { display: inline-block; background: #000; color: #fff; padding: 12px 24px; text-decoration: none; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>ZAVERA</h1>
    </div>
    <div class="content">
        <div class="refund-badge">
            <h2>💰 Refund Berhasil Diproses</h2>
        </div>
        <p>Halo {{.CustomerName}},</p>
        <p>Pengembalian dana untuk pesanan Anda telah berhasil diproses.</p>
        
        <div class="refund-box">
            <p style="margin: 0; color: #666;">Jumlah Refund</p>
            <p class="refund-amount">Rp {{.RefundAmount}}</p>
            <p style="margin: 0; font-size: 14px; color: #666;">{{.RefundMethod}}</p>
        </div>
        
        <div class="order-info">
            <strong>Nomor Pesanan:</strong> {{.OrderCode}}<br>
            <strong>Nomor Refund:</strong> {{.RefundCode}}<br>
            <strong>Tanggal Refund:</strong> {{.RefundedAt}}<br>
            <strong>Alasan:</strong> {{.RefundReason}}
        </div>
        
        <h3>Detail Pesanan yang Di-refund</h3>
        <table class="items-table">
            <tr><th>Produk</th><th>Qty</th><th>Harga</th></tr>
            {{range .Items}}
            <tr><td>{{.ProductName}}</td><td>{{.Quantity}}</td><td>Rp {{.Subtotal}}</td></tr>
            {{end}}
        </table>
        
        <h3>Rincian Refund</h3>
        <div class="order-info">
            <table style="width: 100%;">
                <tr><td>Subtotal Produk</td><td style="text-align: right;">Rp {{.Subtotal}}</td></tr>
                <tr><td>Ongkos Kirim</td><td style="text-align: right;">Rp {{.ShippingCost}}</td></tr>
                <tr style="font-weight: bold; border-top: 1px solid #ddd;">
                    <td>Total Refund</td>
                    <td style="text-align: right; color: #059669;">Rp {{.RefundAmount}}</td>
                </tr>
            </table>
        </div>
        
        <p><strong>Estimasi Waktu:</strong> Dana akan masuk ke rekening/saldo Anda dalam 3-14 hari kerja tergantung metode pembayaran.</p>
        
        <p>Terima kasih atas pengertian Anda. Kami berharap dapat melayani Anda kembali di lain waktu.</p>
        
        <p style="text-align: center; margin: 20px 0;">
            <a href="{{.ShopURL}}" class="btn">Belanja Lagi</a>
        </p>
    </div>
    <div class="footer">
        <p>© 2026 ZAVERA. All rights reserved.</p>
        <p>Jika ada pertanyaan, hubungi kami di support@zavera.com</p>
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
-- VERIFICATION
-- ============================================
SELECT 'Email templates after upgrade:' AS info;
SELECT template_key, name, is_active FROM email_templates ORDER BY template_key;

SELECT 'Email upgrade migration completed!' AS status;
