CREATE TABLE
    IF NOT EXISTS category_color_labels (
        id SERIAL PRIMARY KEY,
        kode_warna TEXT,
        nama_warna TEXT,
        nama_kolom TEXT,
        keterangan TEXT,
        tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );