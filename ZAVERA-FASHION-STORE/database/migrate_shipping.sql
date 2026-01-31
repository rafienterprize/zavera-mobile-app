-- ============================================
-- SHIPPING SYSTEM MIGRATION
-- ZAVERA E-Commerce Shipping Gateway
-- ============================================

-- ============================================
-- SHIPPING PROVIDERS TABLE
-- Stores courier companies (JNE, J&T, SiCepat, etc)
-- ============================================
CREATE TABLE IF NOT EXISTS shipping_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    logo_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_shipping_providers_code ON shipping_providers(code);
CREATE INDEX IF NOT EXISTS idx_shipping_providers_active ON shipping_providers(is_active);

-- ============================================
-- SHIPPING SERVICES TABLE
-- Stores service types per provider (REG, YES, ECO, etc)
-- ============================================
CREATE TABLE IF NOT EXISTS shipping_services (
    id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES shipping_providers(id) ON DELETE CASCADE,
    service_code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    etd VARCHAR(50), -- Estimated time delivery (e.g., "1-2 days")
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, service_code)
);

CREATE INDEX IF NOT EXISTS idx_shipping_services_provider ON shipping_services(provider_id);
CREATE INDEX IF NOT EXISTS idx_shipping_services_code ON shipping_services(service_code);

-- ============================================
-- SHIPMENT STATUS ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'shipment_status') THEN
        CREATE TYPE shipment_status AS ENUM (
            'PENDING',      -- Waiting for payment
            'PROCESSING',   -- Payment received, preparing shipment
            'SHIPPED',      -- Handed to courier
            'IN_TRANSIT',   -- On the way
            'OUT_FOR_DELIVERY', -- Out for delivery
            'DELIVERED',    -- Successfully delivered
            'RETURNED',     -- Returned to sender
            'FAILED'        -- Delivery failed
        );
    END IF;
END $$;

-- ============================================
-- SHIPMENTS TABLE
-- Stores shipment records for each order
-- ============================================
CREATE TABLE IF NOT EXISTS shipments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    
    -- Courier info
    provider_code VARCHAR(50) NOT NULL,
    provider_name VARCHAR(100) NOT NULL,
    service_code VARCHAR(50) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    
    -- Cost & timing
    cost DECIMAL(12, 2) NOT NULL,
    etd VARCHAR(50),
    weight INTEGER NOT NULL, -- in grams
    
    -- Tracking
    tracking_number VARCHAR(100),
    status shipment_status DEFAULT 'PENDING',
    
    -- Origin & destination
    origin_city_id VARCHAR(20),
    origin_city_name VARCHAR(100),
    destination_city_id VARCHAR(20),
    destination_city_name VARCHAR(100),
    
    -- Timestamps
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_shipment_cost CHECK (cost >= 0),
    CONSTRAINT chk_shipment_weight CHECK (weight > 0)
);

CREATE INDEX IF NOT EXISTS idx_shipments_order ON shipments(order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_tracking ON shipments(tracking_number);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments(status);
CREATE INDEX IF NOT EXISTS idx_shipments_provider ON shipments(provider_code);

-- ============================================
-- SHIPMENT TRACKING HISTORY TABLE
-- Stores tracking events from courier API
-- ============================================
CREATE TABLE IF NOT EXISTS shipment_tracking_history (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    status VARCHAR(100) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    event_time TIMESTAMP,
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tracking_history_shipment ON shipment_tracking_history(shipment_id);
CREATE INDEX IF NOT EXISTS idx_tracking_history_time ON shipment_tracking_history(event_time DESC);

-- ============================================
-- ADD WEIGHT TO PRODUCTS TABLE
-- ============================================
ALTER TABLE products ADD COLUMN IF NOT EXISTS weight INTEGER DEFAULT 500; -- default 500 grams

-- ============================================
-- ADD SHIPPING ADDRESS TO ORDERS TABLE
-- ============================================
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_address_id INTEGER;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_address_snapshot JSONB;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_provider_code VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_service_code VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_locked BOOLEAN DEFAULT false;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS total_weight INTEGER DEFAULT 0;

-- ============================================
-- USER ADDRESSES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS user_addresses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(50), -- "Home", "Office", etc
    recipient_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    
    -- Location
    province_id VARCHAR(20),
    province_name VARCHAR(100),
    city_id VARCHAR(20) NOT NULL,
    city_name VARCHAR(100) NOT NULL,
    district VARCHAR(100),
    subdistrict VARCHAR(100),
    postal_code VARCHAR(10),
    full_address TEXT NOT NULL,
    
    -- Flags
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_addresses_user ON user_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_city ON user_addresses(city_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_default ON user_addresses(user_id, is_default) WHERE is_default = true;

-- ============================================
-- SEED SHIPPING PROVIDERS
-- ============================================
INSERT INTO shipping_providers (name, code, logo_url, is_active) VALUES
('JNE Express', 'jne', 'https://www.jne.co.id/frontend/images/logo.png', true),
('J&T Express', 'jnt', 'https://www.jet.co.id/images/logo.png', true),
('SiCepat', 'sicepat', 'https://www.sicepat.com/img/logo.png', true),
('Pos Indonesia', 'pos', 'https://www.posindonesia.co.id/assets/images/logo.png', true),
('TIKI', 'tiki', 'https://www.tiki.id/images/logo.png', true),
('Anteraja', 'anteraja', 'https://anteraja.id/images/logo.png', true),
('Ninja Express', 'ninja', 'https://www.ninjaxpress.co/images/logo.png', true),
('Lion Parcel', 'lion', 'https://lionparcel.com/images/logo.png', true),
('ID Express', 'ide', 'https://idexpress.com/images/logo.png', true),
('SAP Express', 'sap', 'https://sapexpress.co.id/images/logo.png', true)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    logo_url = EXCLUDED.logo_url,
    is_active = EXCLUDED.is_active;

-- ============================================
-- UPDATE TRIGGERS
-- ============================================
DROP TRIGGER IF EXISTS update_shipping_providers_updated_at ON shipping_providers;
CREATE TRIGGER update_shipping_providers_updated_at 
    BEFORE UPDATE ON shipping_providers 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_shipping_services_updated_at ON shipping_services;
CREATE TRIGGER update_shipping_services_updated_at 
    BEFORE UPDATE ON shipping_services 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_shipments_updated_at ON shipments;
CREATE TRIGGER update_shipments_updated_at 
    BEFORE UPDATE ON shipments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_user_addresses_updated_at ON user_addresses;
CREATE TRIGGER update_user_addresses_updated_at 
    BEFORE UPDATE ON user_addresses 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- COMMENTS
-- ============================================
COMMENT ON TABLE shipping_providers IS 'Courier companies available for shipping';
COMMENT ON TABLE shipping_services IS 'Service types offered by each courier';
COMMENT ON TABLE shipments IS 'Shipment records linked to orders';
COMMENT ON TABLE shipment_tracking_history IS 'Tracking events from courier API';
COMMENT ON TABLE user_addresses IS 'Saved addresses for users';
COMMENT ON COLUMN products.weight IS 'Product weight in grams';
COMMENT ON COLUMN orders.shipping_locked IS 'True when shipping is selected and locked for payment';
