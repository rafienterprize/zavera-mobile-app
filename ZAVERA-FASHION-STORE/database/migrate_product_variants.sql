-- Product Variants Migration
-- Adds support for product variants with size, color, and custom attributes

-- Create variant attributes table
CREATE TABLE IF NOT EXISTS variant_attributes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('size', 'color', 'text', 'select')),
    options JSONB,
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create product variants table
CREATE TABLE IF NOT EXISTS product_variants (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL UNIQUE,
    variant_name VARCHAR(255) NOT NULL,
    
    -- Variant attributes
    size VARCHAR(50),
    color VARCHAR(50),
    color_hex VARCHAR(7),
    material VARCHAR(100),
    pattern VARCHAR(100),
    fit VARCHAR(50),
    sleeve VARCHAR(50),
    custom_attributes JSONB,
    
    -- Pricing
    price DECIMAL(10, 2),
    compare_at_price DECIMAL(10, 2),
    cost_per_item DECIMAL(10, 2),
    
    -- Stock
    stock_quantity INT NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    reserved_stock INT NOT NULL DEFAULT 0 CHECK (reserved_stock >= 0),
    low_stock_threshold INT DEFAULT 5,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    
    -- Metadata
    weight_grams INT,
    barcode VARCHAR(100),
    position INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique variant combinations per product
    CONSTRAINT unique_variant_combination UNIQUE (product_id, size, color)
);

-- Create variant images table
CREATE TABLE IF NOT EXISTS variant_images (
    id SERIAL PRIMARY KEY,
    variant_id INT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    alt_text VARCHAR(255),
    position INT DEFAULT 0,
    is_primary BOOLEAN DEFAULT false,
    width INT,
    height INT,
    format VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create stock reservations table
CREATE TABLE IF NOT EXISTS stock_reservations (
    id SERIAL PRIMARY KEY,
    variant_id INT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
    session_id VARCHAR(255),
    quantity INT NOT NULL CHECK (quantity > 0),
    reserved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'expired', 'cancelled')),
    order_id INT REFERENCES orders(id) ON DELETE SET NULL,
    
    INDEX idx_reservations_expires (expires_at, status),
    INDEX idx_reservations_variant (variant_id, status)
);

-- Add variant support to order_items
ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS variant_id INT REFERENCES product_variants(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS variant_sku VARCHAR(100),
ADD COLUMN IF NOT EXISTS variant_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS variant_attributes JSONB;

-- Add variant support to cart_items
ALTER TABLE cart_items
ADD COLUMN IF NOT EXISTS variant_id INT REFERENCES product_variants(id) ON DELETE CASCADE,
ADD COLUMN IF NOT EXISTS variant_sku VARCHAR(100),
ADD COLUMN IF NOT EXISTS variant_attributes JSONB;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_variants_product ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_variants_sku ON product_variants(sku);
CREATE INDEX IF NOT EXISTS idx_variants_active ON product_variants(is_active);
CREATE INDEX IF NOT EXISTS idx_variants_stock ON product_variants(stock_quantity);
CREATE INDEX IF NOT EXISTS idx_variant_images_variant ON variant_images(variant_id);
CREATE INDEX IF NOT EXISTS idx_variant_images_position ON variant_images(variant_id, position);

-- Insert default variant attributes
INSERT INTO variant_attributes (name, display_name, type, options, sort_order) VALUES
('size', 'Size', 'select', '["XS", "S", "M", "L", "XL", "XXL", "XXXL"]'::jsonb, 1),
('color', 'Color', 'color', '[]'::jsonb, 2),
('material', 'Material', 'select', '["Cotton", "Polyester", "Wool", "Silk", "Linen", "Denim", "Leather"]'::jsonb, 3),
('pattern', 'Pattern', 'select', '["Solid", "Striped", "Checked", "Floral", "Geometric", "Abstract"]'::jsonb, 4),
('fit', 'Fit', 'select', '["Slim", "Regular", "Relaxed", "Oversized"]'::jsonb, 5),
('sleeve', 'Sleeve Length', 'select', '["Sleeveless", "Short Sleeve", "3/4 Sleeve", "Long Sleeve"]'::jsonb, 6)
ON CONFLICT (name) DO NOTHING;

-- Function to update product updated_at on variant change
CREATE OR REPLACE FUNCTION update_product_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE products SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.product_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_variant_update_product
AFTER INSERT OR UPDATE ON product_variants
FOR EACH ROW
EXECUTE FUNCTION update_product_timestamp();

-- Function to clean expired reservations
CREATE OR REPLACE FUNCTION clean_expired_reservations()
RETURNS void AS $$
BEGIN
    UPDATE stock_reservations
    SET status = 'expired'
    WHERE status = 'active' AND expires_at < CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql;

-- Function to get available stock for a variant
CREATE OR REPLACE FUNCTION get_available_stock(p_variant_id INT)
RETURNS INT AS $$
DECLARE
    v_stock INT;
    v_reserved INT;
BEGIN
    SELECT stock_quantity INTO v_stock
    FROM product_variants
    WHERE id = p_variant_id;
    
    SELECT COALESCE(SUM(quantity), 0) INTO v_reserved
    FROM stock_reservations
    WHERE variant_id = p_variant_id 
    AND status = 'active' 
    AND expires_at > CURRENT_TIMESTAMP;
    
    RETURN GREATEST(v_stock - v_reserved, 0);
END;
$$ LANGUAGE plpgsql;

-- Function to reserve stock
CREATE OR REPLACE FUNCTION reserve_stock(
    p_variant_id INT,
    p_customer_id INT,
    p_session_id VARCHAR,
    p_quantity INT,
    p_timeout_minutes INT DEFAULT 15
)
RETURNS INT AS $$
DECLARE
    v_available INT;
    v_reservation_id INT;
BEGIN
    -- Clean expired reservations first
    PERFORM clean_expired_reservations();
    
    -- Check available stock
    v_available := get_available_stock(p_variant_id);
    
    IF v_available < p_quantity THEN
        RAISE EXCEPTION 'Insufficient stock. Available: %, Requested: %', v_available, p_quantity;
    END IF;
    
    -- Create reservation
    INSERT INTO stock_reservations (
        variant_id, customer_id, session_id, quantity, expires_at
    ) VALUES (
        p_variant_id, p_customer_id, p_session_id, p_quantity,
        CURRENT_TIMESTAMP + (p_timeout_minutes || ' minutes')::INTERVAL
    ) RETURNING id INTO v_reservation_id;
    
    RETURN v_reservation_id;
END;
$$ LANGUAGE plpgsql;

-- Function to complete reservation (convert to order)
CREATE OR REPLACE FUNCTION complete_reservation(p_reservation_id INT, p_order_id INT)
RETURNS void AS $$
DECLARE
    v_variant_id INT;
    v_quantity INT;
BEGIN
    SELECT variant_id, quantity INTO v_variant_id, v_quantity
    FROM stock_reservations
    WHERE id = p_reservation_id AND status = 'active';
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Reservation not found or already processed';
    END IF;
    
    -- Deduct stock
    UPDATE product_variants
    SET stock_quantity = stock_quantity - v_quantity,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = v_variant_id;
    
    -- Mark reservation as completed
    UPDATE stock_reservations
    SET status = 'completed', order_id = p_order_id
    WHERE id = p_reservation_id;
END;
$$ LANGUAGE plpgsql;

-- Function to cancel reservation
CREATE OR REPLACE FUNCTION cancel_reservation(p_reservation_id INT)
RETURNS void AS $$
BEGIN
    UPDATE stock_reservations
    SET status = 'cancelled'
    WHERE id = p_reservation_id AND status = 'active';
END;
$$ LANGUAGE plpgsql;

-- View for low stock variants
CREATE OR REPLACE VIEW low_stock_variants AS
SELECT 
    pv.id,
    pv.product_id,
    p.name as product_name,
    pv.sku,
    pv.variant_name,
    pv.size,
    pv.color,
    pv.stock_quantity,
    pv.low_stock_threshold,
    get_available_stock(pv.id) as available_stock
FROM product_variants pv
JOIN products p ON p.id = pv.product_id
WHERE pv.is_active = true
AND pv.stock_quantity <= pv.low_stock_threshold
ORDER BY pv.stock_quantity ASC;

-- View for variant stock summary
CREATE OR REPLACE VIEW variant_stock_summary AS
SELECT 
    pv.id as variant_id,
    pv.product_id,
    p.name as product_name,
    pv.sku,
    pv.variant_name,
    pv.stock_quantity,
    COALESCE(SUM(CASE WHEN sr.status = 'active' AND sr.expires_at > CURRENT_TIMESTAMP THEN sr.quantity ELSE 0 END), 0) as reserved_quantity,
    get_available_stock(pv.id) as available_quantity
FROM product_variants pv
JOIN products p ON p.id = pv.product_id
LEFT JOIN stock_reservations sr ON sr.variant_id = pv.id
GROUP BY pv.id, p.id, p.name;
