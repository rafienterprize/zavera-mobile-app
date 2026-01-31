-- Area suggestions table for autocomplete
-- This provides fuzzy search capability that Biteship API doesn't support

CREATE TABLE IF NOT EXISTS area_suggestions (
    id SERIAL PRIMARY KEY,
    area_name VARCHAR(255) NOT NULL,
    full_name VARCHAR(500) NOT NULL,
    province VARCHAR(100),
    city VARCHAR(100),
    district VARCHAR(100),
    postal_code VARCHAR(10),
    biteship_area_id VARCHAR(100),
    search_text VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast text search
CREATE INDEX IF NOT EXISTS idx_area_suggestions_search ON area_suggestions USING gin(to_tsvector('indonesian', search_text));
CREATE INDEX IF NOT EXISTS idx_area_suggestions_name ON area_suggestions(area_name);

-- Insert common Indonesian areas (Jawa Tengah focus for ZAVERA)
INSERT INTO area_suggestions (area_name, full_name, province, city, district, search_text) VALUES
-- Semarang
('Pedurungan', 'Pedurungan, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Pedurungan', 'pedurungan semarang jawa tengah'),
('Semarang Tengah', 'Semarang Tengah, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Semarang Tengah', 'semarang tengah jawa tengah'),
('Semarang Barat', 'Semarang Barat, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Semarang Barat', 'semarang barat jawa tengah'),
('Semarang Timur', 'Semarang Timur, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Semarang Timur', 'semarang timur jawa tengah'),
('Semarang Utara', 'Semarang Utara, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Semarang Utara', 'semarang utara jawa tengah'),
('Semarang Selatan', 'Semarang Selatan, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Semarang Selatan', 'semarang selatan jawa tengah'),
('Tembalang', 'Tembalang, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Tembalang', 'tembalang semarang jawa tengah'),
('Banyumanik', 'Banyumanik, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Banyumanik', 'banyumanik semarang jawa tengah'),
('Candisari', 'Candisari, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Candisari', 'candisari semarang jawa tengah'),
('Gajahmungkur', 'Gajahmungkur, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Gajahmungkur', 'gajahmungkur semarang jawa tengah'),
('Gayamsari', 'Gayamsari, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Gayamsari', 'gayamsari semarang jawa tengah'),
('Genuk', 'Genuk, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Genuk', 'genuk semarang jawa tengah'),
('Gunungpati', 'Gunungpati, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Gunungpati', 'gunungpati semarang jawa tengah'),
('Mijen', 'Mijen, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Mijen', 'mijen semarang jawa tengah'),
('Ngaliyan', 'Ngaliyan, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Ngaliyan', 'ngaliyan semarang jawa tengah'),
('Tugu', 'Tugu, Semarang, Jawa Tengah', 'Jawa Tengah', 'Semarang', 'Tugu', 'tugu semarang jawa tengah'),
-- Jakarta
('Menteng', 'Menteng, Jakarta Pusat, DKI Jakarta', 'DKI Jakarta', 'Jakarta Pusat', 'Menteng', 'menteng jakarta pusat dki jakarta'),
('Kebayoran Baru', 'Kebayoran Baru, Jakarta Selatan, DKI Jakarta', 'DKI Jakarta', 'Jakarta Selatan', 'Kebayoran Baru', 'kebayoran baru jakarta selatan dki jakarta'),
('Kemang', 'Kemang, Jakarta Selatan, DKI Jakarta', 'DKI Jakarta', 'Jakarta Selatan', 'Kemang', 'kemang jakarta selatan dki jakarta'),
('Senayan', 'Senayan, Jakarta Selatan, DKI Jakarta', 'DKI Jakarta', 'Jakarta Selatan', 'Senayan', 'senayan jakarta selatan dki jakarta'),
('Sudirman', 'Sudirman, Jakarta Pusat, DKI Jakarta', 'DKI Jakarta', 'Jakarta Pusat', 'Sudirman', 'sudirman jakarta pusat dki jakarta'),
('Kuningan', 'Kuningan, Jakarta Selatan, DKI Jakarta', 'DKI Jakarta', 'Jakarta Selatan', 'Kuningan', 'kuningan jakarta selatan dki jakarta'),
('Kelapa Gading', 'Kelapa Gading, Jakarta Utara, DKI Jakarta', 'DKI Jakarta', 'Jakarta Utara', 'Kelapa Gading', 'kelapa gading jakarta utara dki jakarta'),
('Sunter', 'Sunter, Jakarta Utara, DKI Jakarta', 'DKI Jakarta', 'Jakarta Utara', 'Sunter', 'sunter jakarta utara dki jakarta'),
('Pluit', 'Pluit, Jakarta Utara, DKI Jakarta', 'DKI Jakarta', 'Jakarta Utara', 'Pluit', 'pluit jakarta utara dki jakarta'),
('Cengkareng', 'Cengkareng, Jakarta Barat, DKI Jakarta', 'DKI Jakarta', 'Jakarta Barat', 'Cengkareng', 'cengkareng jakarta barat dki jakarta')
ON CONFLICT DO NOTHING;
