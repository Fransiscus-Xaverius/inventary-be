-- Seeder for category_color_labels table
-- This seeder adds color labels for different categories

-- Insert or update color labels
INSERT INTO category_color_labels (kode_warna, nama_warna, nama_kolom, keterangan) 
VALUES 
    ('#9b59b6', 'blue',             'kat',    'basic'),
    ('#f1c40f', 'gold',             'kat',    'premium'),
    ('#2ecc71', 'green',            'kat',    'regular'),
    ('#3498db', 'purple',           'gender', 'pria'),
    ('#e91e63', 'pink',             'gender', 'wanita'),
    ('#222222', 'indigo',           'gender', 'unisex'),
    ('#e67e22', 'orange',           'tipe',   'jaket'),
    ('#1abc9c', 'teal',             'tipe',   'topi'),
    ('#34495e', 'dark blue',        'tipe',   'baju'),
    ('#d35400', 'burnt orange',     'tipe',   'sepatu'),
    ('#95a5a6', 'gray',             'tipe',   'celana'),
    ('#2ecc71', 'green',            'status', 'active'),
    ('#e74c3c', 'red',              'status', 'inactive'),
    ('#95a5a6', 'gray',             'status', 'discontinued')
ON CONFLICT (nama_kolom, keterangan) DO UPDATE
SET kode_warna = EXCLUDED.kode_warna,
    nama_warna = EXCLUDED.nama_warna;