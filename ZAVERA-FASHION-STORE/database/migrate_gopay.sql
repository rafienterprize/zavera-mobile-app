-- Migration: Add GoPay support to order_payments table
-- Date: 2026-01-13

-- Add GoPay specific columns
ALTER TABLE order_payments 
ADD COLUMN IF NOT EXISTS qr_code_url TEXT,
ADD COLUMN IF NOT EXISTS deeplink_url TEXT;

-- Add GoPay to payment method enum if not exists
DO $$ 
BEGIN
    -- Check if gopay already exists in the enum
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'gopay' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'va_payment_method')
    ) THEN
        ALTER TYPE va_payment_method ADD VALUE IF NOT EXISTS 'gopay';
    END IF;
END $$;

-- Add GoPay payment instructions
INSERT INTO bank_payment_instructions (bank, channel, step_order, instruction, is_active) VALUES
('gopay', 'Aplikasi Gojek', 1, 'Buka aplikasi Gojek di smartphone Anda', true),
('gopay', 'Aplikasi Gojek', 2, 'Tap menu "Bayar" atau scan QR code', true),
('gopay', 'Aplikasi Gojek', 3, 'Jika scan QR, arahkan kamera ke QR code pembayaran', true),
('gopay', 'Aplikasi Gojek', 4, 'Periksa detail pembayaran dan tap "Konfirmasi & Bayar"', true),
('gopay', 'Aplikasi Gojek', 5, 'Masukkan PIN GoPay Anda', true),
('gopay', 'Aplikasi Gojek', 6, 'Pembayaran selesai', true),
('gopay', 'Aplikasi GoPay', 1, 'Buka aplikasi GoPay di smartphone Anda', true),
('gopay', 'Aplikasi GoPay', 2, 'Tap "Scan QR" atau "Bayar"', true),
('gopay', 'Aplikasi GoPay', 3, 'Scan QR code pembayaran atau masukkan nomor merchant', true),
('gopay', 'Aplikasi GoPay', 4, 'Periksa detail pembayaran', true),
('gopay', 'Aplikasi GoPay', 5, 'Masukkan PIN GoPay Anda', true),
('gopay', 'Aplikasi GoPay', 6, 'Pembayaran selesai', true)
ON CONFLICT DO NOTHING;

-- Create index for faster lookup
CREATE INDEX IF NOT EXISTS idx_order_payments_gopay ON order_payments(payment_method) WHERE payment_method = 'gopay';

COMMENT ON COLUMN order_payments.qr_code_url IS 'QR code URL for GoPay/QRIS payments';
COMMENT ON COLUMN order_payments.deeplink_url IS 'Deeplink URL for GoPay mobile app redirect';
