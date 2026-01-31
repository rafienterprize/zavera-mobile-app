-- ============================================
-- PAYMENT METHOD IMMUTABILITY TRIGGER
-- Prevents modification of payment_method after creation
-- ============================================

-- Trigger function to prevent payment_method modification
CREATE OR REPLACE FUNCTION prevent_payment_method_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Allow if payment_method is not being changed
    IF OLD.payment_method = NEW.payment_method THEN
        RETURN NEW;
    END IF;
    
    -- Prevent any change to payment_method
    RAISE EXCEPTION 'Payment method cannot be modified after creation. Original: %, Attempted: %', 
        OLD.payment_method, NEW.payment_method;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS trg_prevent_payment_method_change ON order_payments;

-- Create trigger
CREATE TRIGGER trg_prevent_payment_method_change
    BEFORE UPDATE ON order_payments
    FOR EACH ROW
    EXECUTE FUNCTION prevent_payment_method_change();

-- Also prevent bank modification (must match payment_method)
CREATE OR REPLACE FUNCTION prevent_bank_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Allow if bank is not being changed
    IF OLD.bank = NEW.bank THEN
        RETURN NEW;
    END IF;
    
    -- Prevent any change to bank
    RAISE EXCEPTION 'Bank cannot be modified after creation. Original: %, Attempted: %', 
        OLD.bank, NEW.bank;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS trg_prevent_bank_change ON order_payments;

-- Create trigger
CREATE TRIGGER trg_prevent_bank_change
    BEFORE UPDATE ON order_payments
    FOR EACH ROW
    EXECUTE FUNCTION prevent_bank_change();

-- Verification
SELECT 'Payment immutability triggers created' AS status;
