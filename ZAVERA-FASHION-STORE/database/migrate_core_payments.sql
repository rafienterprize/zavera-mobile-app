-- ============================================
-- ZAVERA CORE PAYMENT SYSTEM MIGRATION
-- Tokopedia-style payment with Midtrans Core API
-- ============================================

-- Payment method enum for VA types
DO $$ BEGIN
    CREATE TYPE va_payment_method AS ENUM ('bca_va', 'bri_va', 'mandiri_va');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Core payment status enum
DO $$ BEGIN
    CREATE TYPE core_payment_status AS ENUM ('PENDING', 'PAID', 'EXPIRED', 'CANCELLED', 'FAILED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ============================================
-- ORDER_PAYMENTS TABLE
-- Stores VA payment details with immutable payment method
-- ============================================
CREATE TABLE IF NOT EXISTS order_payments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payment_method va_payment_method NOT NULL,
    bank VARCHAR(20) NOT NULL,
    va_number VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(100),
    midtrans_order_id VARCHAR(150) NOT NULL UNIQUE,
    expiry_time TIMESTAMP NOT NULL,
    payment_status core_payment_status NOT NULL DEFAULT 'PENDING',
    raw_response JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,
    
    CONSTRAINT chk_bank_matches_method CHECK (
        (payment_method = 'bca_va' AND bank = 'bca') OR
        (payment_method = 'bri_va' AND bank = 'bri') OR
        (payment_method = 'mandiri_va' AND bank = 'mandiri')
    )
);

-- Partial unique index: Only ONE pending payment per order
-- This enforces the "one active payment per order" rule
CREATE UNIQUE INDEX IF NOT EXISTS idx_order_payments_one_pending 
ON order_payments(order_id) 
WHERE payment_status = 'PENDING';

-- Standard indexes for queries
CREATE INDEX IF NOT EXISTS idx_order_payments_order_id ON order_payments(order_id);
CREATE INDEX IF NOT EXISTS idx_order_payments_status ON order_payments(payment_status);
CREATE INDEX IF NOT EXISTS idx_order_payments_expiry ON order_payments(expiry_time) WHERE payment_status = 'PENDING';
CREATE INDEX IF NOT EXISTS idx_order_payments_midtrans_id ON order_payments(midtrans_order_id);
CREATE INDEX IF NOT EXISTS idx_order_payments_transaction_id ON order_payments(transaction_id);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_order_payments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_order_payments_updated_at ON order_payments;
CREATE TRIGGER update_order_payments_updated_at 
    BEFORE UPDATE ON order_payments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_order_payments_updated_at();

-- ============================================
-- EXTEND ORDER STATUS ENUM
-- Add MENUNGGU_PEMBAYARAN and KADALUARSA if not exists
-- ============================================
DO $$ BEGIN
    -- Check if we need to add new values to order_status enum
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'MENUNGGU_PEMBAYARAN' AND enumtypid = 'order_status'::regtype) THEN
        ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'MENUNGGU_PEMBAYARAN';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'KADALUARSA' AND enumtypid = 'order_status'::regtype) THEN
        ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'KADALUARSA';
    END IF;
EXCEPTION
    WHEN others THEN null;
END $$;

-- ============================================
-- PAYMENT SYNC LOG FOR AUDIT
-- Track all payment state changes
-- ============================================
CREATE TABLE IF NOT EXISTS core_payment_sync_log (
    id SERIAL PRIMARY KEY,
    payment_id INTEGER REFERENCES order_payments(id),
    order_id INTEGER NOT NULL,
    order_code VARCHAR(100) NOT NULL,
    sync_type VARCHAR(20) NOT NULL, -- 'webhook', 'expiry_check', 'manual'
    sync_status VARCHAR(20) NOT NULL, -- 'SYNCED', 'FAILED', 'SKIPPED'
    local_payment_status VARCHAR(20),
    local_order_status VARCHAR(20),
    gateway_status VARCHAR(50),
    gateway_transaction_id VARCHAR(100),
    has_mismatch BOOLEAN DEFAULT false,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_core_payment_sync_log_payment ON core_payment_sync_log(payment_id);
CREATE INDEX IF NOT EXISTS idx_core_payment_sync_log_order ON core_payment_sync_log(order_id);

-- ============================================
-- BANK PAYMENT INSTRUCTIONS
-- Store bank-specific payment instructions
-- ============================================
CREATE TABLE IF NOT EXISTS bank_payment_instructions (
    id SERIAL PRIMARY KEY,
    bank VARCHAR(20) NOT NULL,
    channel VARCHAR(50) NOT NULL, -- 'ATM', 'Mobile Banking', 'Internet Banking'
    step_order INTEGER NOT NULL,
    instruction TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(bank, channel, step_order)
);

-- Insert default BCA instructions
INSERT INTO bank_payment_instructions (bank, channel, step_order, instruction) VALUES
('bca', 'ATM BCA', 1, 'Masukkan kartu ATM dan PIN BCA Anda'),
('bca', 'ATM BCA', 2, 'Pilih menu "Transaksi Lainnya"'),
('bca', 'ATM BCA', 3, 'Pilih menu "Transfer"'),
('bca', 'ATM BCA', 4, 'Pilih menu "Ke Rekening BCA Virtual Account"'),
('bca', 'ATM BCA', 5, 'Masukkan nomor Virtual Account'),
('bca', 'ATM BCA', 6, 'Masukkan jumlah yang ingin dibayarkan'),
('bca', 'ATM BCA', 7, 'Ikuti instruksi untuk menyelesaikan transaksi'),
('bca', 'm-BCA (BCA Mobile)', 1, 'Buka aplikasi BCA Mobile'),
('bca', 'm-BCA (BCA Mobile)', 2, 'Pilih menu "m-Transfer"'),
('bca', 'm-BCA (BCA Mobile)', 3, 'Pilih "BCA Virtual Account"'),
('bca', 'm-BCA (BCA Mobile)', 4, 'Masukkan nomor Virtual Account'),
('bca', 'm-BCA (BCA Mobile)', 5, 'Masukkan PIN m-BCA'),
('bca', 'm-BCA (BCA Mobile)', 6, 'Konfirmasi pembayaran'),
('bca', 'KlikBCA (Internet Banking)', 1, 'Login ke KlikBCA'),
('bca', 'KlikBCA (Internet Banking)', 2, 'Pilih menu "Transfer Dana"'),
('bca', 'KlikBCA (Internet Banking)', 3, 'Pilih "Transfer ke BCA Virtual Account"'),
('bca', 'KlikBCA (Internet Banking)', 4, 'Masukkan nomor Virtual Account'),
('bca', 'KlikBCA (Internet Banking)', 5, 'Masukkan jumlah pembayaran'),
('bca', 'KlikBCA (Internet Banking)', 6, 'Masukkan respon KeyBCA Appli 1'),
('bca', 'KlikBCA (Internet Banking)', 7, 'Konfirmasi pembayaran')
ON CONFLICT (bank, channel, step_order) DO NOTHING;

-- Insert default BRI instructions
INSERT INTO bank_payment_instructions (bank, channel, step_order, instruction) VALUES
('bri', 'ATM BRI', 1, 'Masukkan kartu ATM dan PIN BRI Anda'),
('bri', 'ATM BRI', 2, 'Pilih menu "Transaksi Lain"'),
('bri', 'ATM BRI', 3, 'Pilih menu "Pembayaran"'),
('bri', 'ATM BRI', 4, 'Pilih menu "Lainnya"'),
('bri', 'ATM BRI', 5, 'Pilih menu "BRIVA"'),
('bri', 'ATM BRI', 6, 'Masukkan nomor Virtual Account'),
('bri', 'ATM BRI', 7, 'Konfirmasi pembayaran'),
('bri', 'BRImo (Mobile Banking)', 1, 'Buka aplikasi BRImo'),
('bri', 'BRImo (Mobile Banking)', 2, 'Pilih menu "BRIVA"'),
('bri', 'BRImo (Mobile Banking)', 3, 'Masukkan nomor Virtual Account'),
('bri', 'BRImo (Mobile Banking)', 4, 'Masukkan PIN BRImo'),
('bri', 'BRImo (Mobile Banking)', 5, 'Konfirmasi pembayaran'),
('bri', 'Internet Banking BRI', 1, 'Login ke Internet Banking BRI'),
('bri', 'Internet Banking BRI', 2, 'Pilih menu "Pembayaran"'),
('bri', 'Internet Banking BRI', 3, 'Pilih "BRIVA"'),
('bri', 'Internet Banking BRI', 4, 'Masukkan nomor Virtual Account'),
('bri', 'Internet Banking BRI', 5, 'Masukkan password dan mToken'),
('bri', 'Internet Banking BRI', 6, 'Konfirmasi pembayaran')
ON CONFLICT (bank, channel, step_order) DO NOTHING;

-- Insert default Mandiri instructions
INSERT INTO bank_payment_instructions (bank, channel, step_order, instruction) VALUES
('mandiri', 'ATM Mandiri', 1, 'Masukkan kartu ATM dan PIN Mandiri Anda'),
('mandiri', 'ATM Mandiri', 2, 'Pilih menu "Bayar/Beli"'),
('mandiri', 'ATM Mandiri', 3, 'Pilih menu "Lainnya"'),
('mandiri', 'ATM Mandiri', 4, 'Pilih menu "Multi Payment"'),
('mandiri', 'ATM Mandiri', 5, 'Masukkan kode perusahaan: 70012'),
('mandiri', 'ATM Mandiri', 6, 'Masukkan nomor Virtual Account'),
('mandiri', 'ATM Mandiri', 7, 'Konfirmasi pembayaran'),
('mandiri', 'Livin by Mandiri (Mobile)', 1, 'Buka aplikasi Livin by Mandiri'),
('mandiri', 'Livin by Mandiri (Mobile)', 2, 'Pilih menu "Bayar"'),
('mandiri', 'Livin by Mandiri (Mobile)', 3, 'Pilih "Multipayment"'),
('mandiri', 'Livin by Mandiri (Mobile)', 4, 'Pilih penyedia jasa: Midtrans'),
('mandiri', 'Livin by Mandiri (Mobile)', 5, 'Masukkan nomor Virtual Account'),
('mandiri', 'Livin by Mandiri (Mobile)', 6, 'Masukkan PIN Livin'),
('mandiri', 'Livin by Mandiri (Mobile)', 7, 'Konfirmasi pembayaran'),
('mandiri', 'Internet Banking Mandiri', 1, 'Login ke Mandiri Online'),
('mandiri', 'Internet Banking Mandiri', 2, 'Pilih menu "Pembayaran"'),
('mandiri', 'Internet Banking Mandiri', 3, 'Pilih "Multi Payment"'),
('mandiri', 'Internet Banking Mandiri', 4, 'Pilih penyedia jasa: Midtrans'),
('mandiri', 'Internet Banking Mandiri', 5, 'Masukkan nomor Virtual Account'),
('mandiri', 'Internet Banking Mandiri', 6, 'Masukkan token'),
('mandiri', 'Internet Banking Mandiri', 7, 'Konfirmasi pembayaran')
ON CONFLICT (bank, channel, step_order) DO NOTHING;

-- ============================================
-- VERIFICATION QUERY
-- ============================================
-- SELECT 'order_payments table created' AS status WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'order_payments');
-- SELECT 'Partial unique index created' AS status WHERE EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_order_payments_one_pending');
