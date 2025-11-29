-- Seeder for master_colors table with Indonesian color names and hex values

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS master_colors (
    id SERIAL PRIMARY KEY,
    nama TEXT NOT NULL,
    hex TEXT,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tanggal_hapus TIMESTAMPTZ
);

-- Create index on nama field
CREATE INDEX IF NOT EXISTS idx_master_colors_nama ON master_colors(nama);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_colors_nama'
    ) THEN
        ALTER TABLE master_colors ADD CONSTRAINT uq_master_colors_nama UNIQUE (nama);
    END IF;
END$$;

-- Truncate the table first (optional, only if you want to start fresh)
-- TRUNCATE TABLE master_colors RESTART IDENTITY;

-- Insert colors with hex values
INSERT INTO master_colors (nama, hex, tanggal_update) VALUES
('Merah', '#FF0000', CURRENT_TIMESTAMP),
('Biru', '#0000FF', CURRENT_TIMESTAMP),
('Hijau', '#00FF00', CURRENT_TIMESTAMP),
('Kuning', '#FFFF00', CURRENT_TIMESTAMP),
('Hitam', '#000000', CURRENT_TIMESTAMP),
('Putih', '#FFFFFF', CURRENT_TIMESTAMP),
('Abu-abu', '#808080', CURRENT_TIMESTAMP),
('Ungu', '#800080', CURRENT_TIMESTAMP),
('Jingga', '#FFA500', CURRENT_TIMESTAMP),
('Merah Muda', '#FFC0CB', CURRENT_TIMESTAMP),
('Coklat', '#A52A2A', CURRENT_TIMESTAMP),
('Biru Muda', '#ADD8E6', CURRENT_TIMESTAMP),
('Biru Tua', '#00008B', CURRENT_TIMESTAMP),
('Hijau Muda', '#90EE90', CURRENT_TIMESTAMP),
('Hijau Tua', '#006400', CURRENT_TIMESTAMP),
('Merah Tua', '#8B0000', CURRENT_TIMESTAMP),
('Emas', '#FFD700', CURRENT_TIMESTAMP),
('Perak', '#C0C0C0', CURRENT_TIMESTAMP),
('Krem', '#FFFDD0', CURRENT_TIMESTAMP),
('Magenta', '#FF00FF', CURRENT_TIMESTAMP),
('Cyan', '#00FFFF', CURRENT_TIMESTAMP),
('Marun', '#800000', CURRENT_TIMESTAMP),
('Navy', '#000080', CURRENT_TIMESTAMP),
('Olive', '#808000', CURRENT_TIMESTAMP),
('Teal', '#008080', CURRENT_TIMESTAMP)
ON CONFLICT (nama) DO UPDATE
SET hex = EXCLUDED.hex, tanggal_update = CURRENT_TIMESTAMP;

-- Remove unique constraint after insertions
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_colors_nama'
    ) THEN
        ALTER TABLE master_colors DROP CONSTRAINT uq_master_colors_nama;
    END IF;
END$$;