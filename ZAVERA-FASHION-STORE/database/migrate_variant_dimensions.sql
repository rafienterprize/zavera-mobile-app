-- Add shipping dimensions to product_variants table
-- These are used for accurate shipping cost calculation

ALTER TABLE product_variants 
ADD COLUMN IF NOT EXISTS length_cm INT,
ADD COLUMN IF NOT EXISTS width_cm INT,
ADD COLUMN IF NOT EXISTS height_cm INT;

COMMENT ON COLUMN product_variants.weight_grams IS 'Weight in grams for shipping calculation';
COMMENT ON COLUMN product_variants.length_cm IS 'Length in cm - used for volumetric weight (max of all items)';
COMMENT ON COLUMN product_variants.width_cm IS 'Width in cm - used for volumetric weight (max of all items)';
COMMENT ON COLUMN product_variants.height_cm IS 'Height in cm - used for volumetric weight (sum of all items when stacked)';

-- Create index for shipping calculations
CREATE INDEX IF NOT EXISTS idx_variants_dimensions ON product_variants(weight_grams, length_cm, width_cm, height_cm) 
WHERE weight_grams IS NOT NULL;

-- Verify the changes
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'product_variants' 
AND column_name IN ('weight_grams', 'length_cm', 'width_cm', 'height_cm');
