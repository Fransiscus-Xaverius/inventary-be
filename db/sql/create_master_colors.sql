CREATE TABLE
    IF NOT EXISTS master_colors (
        id SERIAL PRIMARY KEY,
        nama TEXT,
        tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        tanggal_hapus TIMESTAMPTZ DEFAULT NULL,
    );