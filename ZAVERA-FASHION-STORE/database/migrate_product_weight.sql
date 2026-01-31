-- Migration: Add weight column to products table (PostgreSQL)
-- This fixes the shipping weight calculation bug where hardcoded 500g was used

-- Add weight column (in grams) if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'products' AND column_name = 'weight') THEN
        ALTER TABLE products ADD COLUMN weight INT DEFAULT 500;
    END IF;
END $$;

-- Update existing products with realistic weights for fashion items
-- T-shirts, shirts: 200-350g
UPDATE products SET weight = 350 WHERE LOWER(name) LIKE '%kaos%' OR LOWER(name) LIKE '%t-shirt%' OR LOWER(name) LIKE '%tshirt%';
UPDATE products SET weight = 400 WHERE LOWER(name) LIKE '%kemeja%' OR LOWER(name) LIKE '%shirt%';

-- Jackets, hoodies: 500-800g
UPDATE products SET weight = 600 WHERE LOWER(name) LIKE '%jaket%' OR LOWER(name) LIKE '%jacket%';
UPDATE products SET weight = 700 WHERE LOWER(name) LIKE '%hoodie%' OR LOWER(name) LIKE '%sweater%';

-- Pants, jeans: 400-600g
UPDATE products SET weight = 500 WHERE LOWER(name) LIKE '%celana%' OR LOWER(name) LIKE '%pants%' OR LOWER(name) LIKE '%jeans%';

-- Dresses, skirts: 300-500g
UPDATE products SET weight = 400 WHERE LOWER(name) LIKE '%dress%' OR LOWER(name) LIKE '%rok%' OR LOWER(name) LIKE '%skirt%';

-- Accessories: 100-200g
UPDATE products SET weight = 150 WHERE LOWER(name) LIKE '%topi%' OR LOWER(name) LIKE '%hat%' OR LOWER(name) LIKE '%cap%';
UPDATE products SET weight = 100 WHERE LOWER(name) LIKE '%syal%' OR LOWER(name) LIKE '%scarf%';

-- Verify the update
SELECT id, name, weight FROM products ORDER BY id;
