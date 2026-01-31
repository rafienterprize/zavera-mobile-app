-- ============================================
-- ZAVERA E-COMMERCE DATABASE SCHEMA
-- PostgreSQL 12+
-- ============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- USERS TABLE
-- ============================================
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    password_hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- ============================================
-- PRODUCTS TABLE
-- ============================================
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    price DECIMAL(12, 2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_price_positive CHECK (price >= 0),
    CONSTRAINT chk_stock_non_negative CHECK (stock >= 0)
);

CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_active ON products(is_active);
CREATE INDEX idx_products_created ON products(created_at DESC);

-- ============================================
-- PRODUCT IMAGES TABLE
-- ============================================
CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url VARCHAR(500) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_product_images_product ON product_images(product_id);
CREATE INDEX idx_product_images_primary ON product_images(product_id, is_primary) WHERE is_primary = true;

-- ============================================
-- CARTS TABLE
-- ============================================
CREATE TABLE carts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_cart_identity CHECK (user_id IS NOT NULL OR session_id IS NOT NULL)
);

CREATE INDEX idx_carts_user ON carts(user_id);
CREATE INDEX idx_carts_session ON carts(session_id);

-- ============================================
-- CART ITEMS TABLE
-- ============================================
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    price_snapshot DECIMAL(12, 2) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_quantity_positive CHECK (quantity > 0),
    UNIQUE(cart_id, product_id)
);

CREATE INDEX idx_cart_items_cart ON cart_items(cart_id);
CREATE INDEX idx_cart_items_product ON cart_items(product_id);

-- ============================================
-- ORDER STATUS ENUM
-- ============================================
CREATE TYPE order_status AS ENUM ('PENDING', 'PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED', 'FAILED');

-- ============================================
-- ORDERS TABLE
-- ============================================
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_code VARCHAR(100) UNIQUE NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(50) NOT NULL,
    
    -- Pricing
    subtotal DECIMAL(12, 2) NOT NULL,
    shipping_cost DECIMAL(12, 2) DEFAULT 0,
    tax DECIMAL(12, 2) DEFAULT 0,
    discount DECIMAL(12, 2) DEFAULT 0,
    total_amount DECIMAL(12, 2) NOT NULL,
    
    -- Status
    status order_status DEFAULT 'PENDING',
    
    -- Metadata
    notes TEXT,
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    
    CONSTRAINT chk_amounts_positive CHECK (
        subtotal >= 0 AND 
        shipping_cost >= 0 AND 
        tax >= 0 AND 
        discount >= 0 AND 
        total_amount >= 0
    )
);

CREATE INDEX idx_orders_code ON orders(order_code);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created_at DESC);
CREATE INDEX idx_orders_email ON orders(customer_email);

-- ============================================
-- ORDER ITEMS TABLE
-- ============================================
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    price_per_unit DECIMAL(12, 2) NOT NULL,
    subtotal DECIMAL(12, 2) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_order_item_quantity CHECK (quantity > 0),
    CONSTRAINT chk_order_item_prices CHECK (price_per_unit >= 0 AND subtotal >= 0)
);

CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_product ON order_items(product_id);

-- ============================================
-- PAYMENT STATUS ENUM
-- ============================================
CREATE TYPE payment_status AS ENUM ('PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'EXPIRED', 'CANCELLED');

-- ============================================
-- PAYMENTS TABLE
-- ============================================
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payment_method VARCHAR(50),
    payment_provider VARCHAR(100),
    
    -- Amount
    amount DECIMAL(12, 2) NOT NULL,
    
    -- Status
    status payment_status DEFAULT 'PENDING',
    
    -- External reference
    external_id VARCHAR(255),
    transaction_id VARCHAR(255),
    
    -- Response data
    provider_response JSONB,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,
    expired_at TIMESTAMP,
    
    CONSTRAINT chk_payment_amount CHECK (amount >= 0)
);

CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_external ON payments(external_id);
CREATE INDEX idx_payments_transaction ON payments(transaction_id);

-- ============================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_carts_updated_at BEFORE UPDATE ON carts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_cart_items_updated_at BEFORE UPDATE ON cart_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- SEED SAMPLE PRODUCTS
-- ============================================
INSERT INTO products (name, slug, description, price, stock, is_active) VALUES
('Minimalist Cotton Tee', 'minimalist-cotton-tee', 'Premium cotton t-shirt with a modern minimalist design. Perfect for everyday wear.', 299000, 50, true),
('Classic Denim Jacket', 'classic-denim-jacket', 'Timeless denim jacket crafted from high-quality denim. A wardrobe essential.', 899000, 30, true),
('Tailored Trousers', 'tailored-trousers', 'Elegant tailored trousers with a perfect fit. Suitable for both casual and formal occasions.', 749000, 40, true),
('Premium Hoodie', 'premium-hoodie', 'Comfortable premium hoodie made from soft cotton blend. Perfect for layering.', 599000, 60, true),
('Slim Fit Shirt', 'slim-fit-shirt', 'Modern slim fit shirt in premium fabric. Ideal for professional settings.', 449000, 45, true),
('Casual Blazer', 'casual-blazer', 'Sophisticated casual blazer with contemporary styling. Elevate any outfit.', 1299000, 25, true),
('Knit Sweater', 'knit-sweater', 'Cozy knit sweater in neutral tones. Perfect for cooler weather.', 549000, 55, true),
('Relaxed Fit Pants', 'relaxed-fit-pants', 'Comfortable relaxed fit pants with modern aesthetics. Great for casual wear.', 649000, 35, true);

-- Add product images
INSERT INTO product_images (product_id, image_url, is_primary, display_order) VALUES
(1, 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=800&q=80', true, 1),
(2, 'https://images.unsplash.com/photo-1551028719-00167b16eac5?w=800&q=80', true, 1),
(3, 'https://images.unsplash.com/photo-1624378439575-d8705ad7ae80?w=800&q=80', true, 1),
(4, 'https://images.unsplash.com/photo-1556821840-3a63f95609a7?w=800&q=80', true, 1),
(5, 'https://images.unsplash.com/photo-1602810318383-e386cc2a3ccf?w=800&q=80', true, 1),
(6, 'https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=800&q=80', true, 1),
(7, 'https://images.unsplash.com/photo-1576566588028-4147f3842f27?w=800&q=80', true, 1),
(8, 'https://images.unsplash.com/photo-1473966968600-fa801b869a1a?w=800&q=80', true, 1);
