CREATE TABLE
    IF NOT EXISTS master_products (
        no SERIAL PRIMARY KEY,
        artikel TEXT NOT NULL,
        nama TEXT,
        deskripsi TEXT,
        rating JSONB,
        warna TEXT,
        size TEXT,
        grup TEXT,
        unit TEXT,
        kat TEXT,
        model TEXT,
        gender TEXT,
        tipe TEXT,
        harga NUMERIC(15, 2),
        harga_diskon NUMERIC(15, 2),
        marketplace JSONB,
        offline JSONB,
        gambar TEXT[],
        tanggal_produk DATE,
        tanggal_terima DATE,
        usia INT,
        status TEXT,
        supplier TEXT,
        diupdate_oleh TEXT,
        tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        tanggal_hapus TIMESTAMPTZ DEFAULT NULL
    );

-- Create index for offline field
CREATE INDEX IF NOT EXISTS idx_master_products_offline ON master_products USING GIN (offline);