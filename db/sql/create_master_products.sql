CREATE TABLE
    IF NOT EXISTS master_products (
        no SERIAL PRIMARY KEY,
        artikel TEXT NOT NULL,
        warna TEXT,
        size TEXT,
        grup TEXT,
        unit TEXT,
        kat TEXT,
        model TEXT,
        gender TEXT,
        tipe TEXT,
        harga NUMERIC(15, 2),
        tanggal_produk DATE,
        tanggal_terima DATE,
        usia INT,
        status TEXT,
        supplier TEXT,
        diupdate_oleh TEXT,
        tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        tanggal_hapus TIMESTAMPTZ DEFAULT NULL,
    );