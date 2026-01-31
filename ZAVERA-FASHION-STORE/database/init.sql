-- Create database (run this separately as superuser)
-- CREATE DATABASE zavera_db;

-- Connect to the database before running the rest
-- \c zavera_db;

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    description TEXT,
    image_url VARCHAR(500),
    stock INTEGER NOT NULL DEFAULT 0
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order_code VARCHAR(100) UNIQUE NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    total_amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price NUMERIC(10, 2) NOT NULL
);

-- Seed sample products
INSERT INTO products (name, price, description, image_url, stock) VALUES
('Minimalist Cotton Tee', 299000, 'Premium cotton t-shirt with a modern minimalist design. Perfect for everyday wear.', 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=800&q=80', 50),
('Classic Denim Jacket', 899000, 'Timeless denim jacket crafted from high-quality denim. A wardrobe essential.', 'https://images.unsplash.com/photo-1551028719-00167b16eac5?w=800&q=80', 30),
('Tailored Trousers', 749000, 'Elegant tailored trousers with a perfect fit. Suitable for both casual and formal occasions.', 'https://images.unsplash.com/photo-1624378439575-d8705ad7ae80?w=800&q=80', 40),
('Premium Hoodie', 599000, 'Comfortable premium hoodie made from soft cotton blend. Perfect for layering.', 'https://images.unsplash.com/photo-1556821840-3a63f95609a7?w=800&q=80', 60),
('Slim Fit Shirt', 449000, 'Modern slim fit shirt in premium fabric. Ideal for professional settings.', 'https://images.unsplash.com/photo-1602810318383-e386cc2a3ccf?w=800&q=80', 45),
('Casual Blazer', 1299000, 'Sophisticated casual blazer with contemporary styling. Elevate any outfit.', 'https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=800&q=80', 25),
('Knit Sweater', 549000, 'Cozy knit sweater in neutral tones. Perfect for cooler weather.', 'https://images.unsplash.com/photo-1576566588028-4147f3842f27?w=800&q=80', 55),
('Relaxed Fit Pants', 649000, 'Comfortable relaxed fit pants with modern aesthetics. Great for casual wear.', 'https://images.unsplash.com/photo-1473966968600-fa801b869a1a?w=800&q=80', 35);
