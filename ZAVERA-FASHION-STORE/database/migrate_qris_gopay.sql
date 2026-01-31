-- Migration: Add QRIS and GoPay to va_payment_method enum
-- Run this migration to enable QRIS and GoPay payment methods

-- Add 'qris' to va_payment_method enum if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'qris' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'va_payment_method')
    ) THEN
        ALTER TYPE va_payment_method ADD VALUE 'qris';
    END IF;
END$$;

-- Add 'gopay' to va_payment_method enum if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'gopay' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'va_payment_method')
    ) THEN
        ALTER TYPE va_payment_method ADD VALUE 'gopay';
    END IF;
END$$;

-- Add 'credit_card' to va_payment_method enum if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'credit_card' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'va_payment_method')
    ) THEN
        ALTER TYPE va_payment_method ADD VALUE 'credit_card';
    END IF;
END$$;

-- Verify the enum values
SELECT enumlabel FROM pg_enum 
WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = 'va_payment_method')
ORDER BY enumsortorder;
