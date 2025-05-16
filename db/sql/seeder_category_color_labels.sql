-- Seeder for category_color_labels table
-- This seeder adds color labels for different categories

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS category_color_labels (
    id SERIAL PRIMARY KEY,
    kode_warna TEXT,
    nama_warna TEXT,
    nama_kolom TEXT,
    keterangan TEXT,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_category_color_labels_column_desc'
    ) THEN
        ALTER TABLE category_color_labels ADD CONSTRAINT uq_category_color_labels_column_desc UNIQUE (nama_kolom, keterangan);
    END IF;
END$$;

-- Insert or update color labels
INSERT INTO category_color_labels (kode_warna, nama_warna, nama_kolom, keterangan) 
VALUES 
    ('#d3d3d3', 'light gray',               'grup',     'casual'),
    ('#add8e6', 'light blue',               'grup',     'formal'),
    ('#90ee90', 'neon greenish yellow',     'grup',     'sport'),
    ('#ff6b6b', 'bright red',               'grup',     'kids'),
    ('#f4d03f', 'dull yellow',              'grup',     'accessories'),
    ('#ffffff', 'white',                    'unit',     'pcs'),
    ('#ffffff', 'white',                    'unit',     'box'),
    ('#ffffff', 'white',                    'unit',     'set'),
    ('#9b59b6', 'blue',                     'kat',      'basic'),
    ('#f1c40f', 'gold',                     'kat',      'premium'),
    ('#2ecc71', 'green',                    'kat',      'regular'),
    ('#3498db', 'purple',                   'gender',   'pria'),
    ('#e91e63', 'pink',                     'gender',   'wanita'),
    ('#222222', 'indigo',                   'gender',   'unisex'),
    ('#e67e22', 'orange',                   'tipe',     'jaket'),
    ('#1abc9c', 'teal',                     'tipe',     'topi'),
    ('#34495e', 'dark blue',                'tipe',     'baju'),
    ('#d35400', 'burnt orange',             'tipe',     'sepatu'),
    ('#95a5a6', 'gray',                     'tipe',     'celana'),
    ('#2ecc71', 'green',                    'status',   'active'),
    ('#e74c3c', 'red',                      'status',   'inactive'),
    ('#95a5a6', 'gray',                     'status',   'discontinued'),
    ('#90ee90', 'light green',              'usia',     'fresh'),
    ('#daa520', 'gold',                     'usia',     'normal'),
    ('#8b0000', 'dark red',                 'usia',     'aging')
ON CONFLICT (nama_kolom, keterangan) DO UPDATE
SET kode_warna = EXCLUDED.kode_warna,
    nama_warna = EXCLUDED.nama_warna;