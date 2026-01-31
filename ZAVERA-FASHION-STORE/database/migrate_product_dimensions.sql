-- ============================================
-- ADD PRODUCT DIMENSIONS FOR BITESHIP
-- ============================================
-- Add length, width, height columns to products table
-- Required for accurate Biteship shipping cost calculation
-- ============================================

-- Add dimension columns (in centimeters)
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS length INTEGER DEFAULT 10,
ADD COLUMN IF NOT EXISTS width INTEGER DEFAULT 10,
ADD COLUMN IF NOT EXISTS height INTEGER DEFAULT 5;

-- Add constraints
ALTER TABLE products
ADD CONSTRAINT chk_length_positive CHECK (length > 0),
ADD CONSTRAINT chk_width_positive CHECK (width > 0),
ADD CONSTRAINT chk_height_positive CHECK (height > 0);

-- Add comments
COMMENT ON COLUMN products.length IS 'Product length in centimeters (for shipping calculation)';
COMMENT ON COLUMN products.width IS 'Product width in centimeters (for shipping calculation)';
COMMENT ON COLUMN products.height IS 'Product height in centimeters (for shipping calculation)';
COMMENT ON COLUMN products.weight IS 'Product weight in grams (for shipping calculation)';

-- Update existing products with default dimensions
UPDATE products 
SET length = 30, width = 20, height = 5
WHERE length IS NULL OR width IS NULL OR height IS NULL;

SELECT 'Product dimensions migration completed successfully' AS status;
