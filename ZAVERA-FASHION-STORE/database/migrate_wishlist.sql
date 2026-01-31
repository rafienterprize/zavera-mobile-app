-- Wishlist Feature Migration
-- Create wishlist table for storing user's favorite products

CREATE TABLE IF NOT EXISTS wishlists (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure user can only add same product once
    UNIQUE(user_id, product_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_wishlists_user_id ON wishlists(user_id);
CREATE INDEX IF NOT EXISTS idx_wishlists_product_id ON wishlists(product_id);
CREATE INDEX IF NOT EXISTS idx_wishlists_created_at ON wishlists(created_at DESC);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_wishlists_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_wishlists_updated_at
    BEFORE UPDATE ON wishlists
    FOR EACH ROW
    EXECUTE FUNCTION update_wishlists_updated_at();

-- Add wishlist_count to products table for quick access
ALTER TABLE products ADD COLUMN IF NOT EXISTS wishlist_count INTEGER DEFAULT 0;

-- Function to update product wishlist count
CREATE OR REPLACE FUNCTION update_product_wishlist_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE products SET wishlist_count = wishlist_count + 1 WHERE id = NEW.product_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE products SET wishlist_count = GREATEST(wishlist_count - 1, 0) WHERE id = OLD.product_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_product_wishlist_count
    AFTER INSERT OR DELETE ON wishlists
    FOR EACH ROW
    EXECUTE FUNCTION update_product_wishlist_count();

COMMENT ON TABLE wishlists IS 'User wishlist/favorites - stores products users want to save for later';
COMMENT ON COLUMN wishlists.user_id IS 'User who added the product to wishlist';
COMMENT ON COLUMN wishlists.product_id IS 'Product added to wishlist';
