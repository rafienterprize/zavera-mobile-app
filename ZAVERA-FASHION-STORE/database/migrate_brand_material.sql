-- Migration: Add brand and material columns to products table
-- Date: 2026-01-30
-- Description: Add brand and material fields to support product details

-- Add brand column
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS brand VARCHAR(100) DEFAULT '';

-- Add material column
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS material VARCHAR(100) DEFAULT '';

-- Add comments for documentation
COMMENT ON COLUMN products.brand IS 'Product brand name (e.g., Nike, Adidas, Zara)';
COMMENT ON COLUMN products.material IS 'Product material (e.g., Cotton, Polyester, Wool)';

-- Create index for faster brand filtering (optional but recommended)
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_material ON products(material);

-- Verify the columns were added
SELECT column_name, data_type, character_maximum_length, column_default
FROM information_schema.columns
WHERE table_name = 'products' 
AND column_name IN ('brand', 'material');
