-- Seeder for master_products table
-- This seeder will add 1000 sample products

-- Set up the seeder
DO $$
DECLARE
    color_names TEXT[];
    color_count INTEGER;
BEGIN
    -- Get all color names from master_colors table
    SELECT ARRAY_AGG(nama) INTO color_names FROM master_colors WHERE tanggal_hapus IS NULL;
    
    -- If no colors found, use default colors
    IF color_names IS NULL OR array_length(color_names, 1) IS NULL THEN
        color_names := ARRAY['Merah', 'Biru', 'Hitam', 'Putih', 'Abu-abu', 'Hijau', 'Kuning'];
    END IF;
    
    -- Get count of colors for random selection
    color_count := array_length(color_names, 1);
    
    -- Insert sample products
    INSERT INTO master_products (artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh, tanggal_update)
    SELECT 
        'ART-' || LPAD(CAST(generate_series AS TEXT), 6, '0') as artikel,
        color_names[floor(random()*color_count + 1)] as warna,
        (ARRAY['S', 'M', 'L', 'XL', 'XXL'])[floor(random()*5 + 1)] as size,
        (ARRAY['Casual', 'Formal', 'Sport', 'Kids', 'Accessories'])[floor(random()*5 + 1)] as grup,
        (ARRAY['PCS', 'BOX', 'SET'])[floor(random()*3 + 1)] as unit,
        (ARRAY['Premium', 'Regular', 'Basic'])[floor(random()*3 + 1)] as kat,
        'MODEL-' || floor(random()*100 + 1) as model,
        (ARRAY['Pria', 'Wanita', 'Unisex'])[floor(random()*3 + 1)] as gender,
        (ARRAY['Baju', 'Celana', 'Jaket', 'Topi', 'Sepatu'])[floor(random()*5 + 1)] as tipe,
        floor(random()*(1000000-50000 + 1) + 50000)::numeric(15,2) as harga,
        CURRENT_DATE - (floor(random()*365)::int || ' days')::interval as tanggal_produk,
        CURRENT_DATE - (floor(random()*180)::int || ' days')::interval as tanggal_terima,
        floor(random()*365 + 1)::int as usia,
        (ARRAY['Active', 'Inactive', 'Discontinued'])[floor(random()*3 + 1)] as status,
        (ARRAY['Supplier A', 'Supplier B', 'Supplier C', 'Supplier D', 'Supplier E'])[floor(random()*5 + 1)] as supplier,
        (ARRAY['Admin1', 'Admin2', 'Admin3', 'System'])[floor(random()*4 + 1)] as diupdate_oleh,
        CURRENT_TIMESTAMP as tanggal_update
    FROM generate_series(1, 1000)
    ON CONFLICT (artikel) DO UPDATE
    SET 
        warna = EXCLUDED.warna,
        size = EXCLUDED.size,
        grup = EXCLUDED.grup,
        unit = EXCLUDED.unit,
        kat = EXCLUDED.kat,
        model = EXCLUDED.model,
        gender = EXCLUDED.gender,
        tipe = EXCLUDED.tipe,
        harga = EXCLUDED.harga,
        tanggal_produk = EXCLUDED.tanggal_produk,
        tanggal_terima = EXCLUDED.tanggal_terima,
        usia = EXCLUDED.usia,
        status = EXCLUDED.status,
        supplier = EXCLUDED.supplier,
        diupdate_oleh = EXCLUDED.diupdate_oleh,
        tanggal_update = CURRENT_TIMESTAMP;
        
    RAISE NOTICE 'Added/updated sample products in the master_products table';
END $$;