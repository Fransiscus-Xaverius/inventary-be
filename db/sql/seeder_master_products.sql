-- Seeder for master_products table
-- This seeder will add 1000 sample products

-- Set up the seeder
DO $$
DECLARE
    color_names TEXT[];
    color_count INTEGER;
    size_values TEXT[];
    size_units TEXT[];
    size_pairs TEXT[];
    size_count INTEGER;
    random_sizes TEXT;
    i INTEGER;
    product_id INTEGER;
    size_list TEXT := '';
    num_sizes INTEGER;
BEGIN
    -- Get all color names from master_colors table
    SELECT ARRAY_AGG(nama) INTO color_names FROM master_colors WHERE tanggal_hapus IS NULL;
    
    -- If no colors found, use default colors
    IF color_names IS NULL OR array_length(color_names, 1) IS NULL THEN
        color_names := ARRAY['Merah', 'Biru', 'Hitam', 'Putih', 'Abu-abu', 'Hijau', 'Kuning'];
    END IF;
    
    -- Get count of colors for random selection
    color_count := array_length(color_names, 1);

    -- Get all size values and units from master_sizes table
    SELECT 
        ARRAY_AGG(value),
        ARRAY_AGG(
            CASE
                WHEN unit = 'US Men' THEN 'USM'
                WHEN unit = 'US Women' THEN 'USW'
                ELSE unit
            END
        )
    INTO 
        size_values,
        size_units
    FROM master_sizes 
    WHERE tanggal_hapus IS NULL;
    
    -- If no sizes found, use default sizes with units
    IF size_values IS NULL OR array_length(size_values, 1) IS NULL THEN
        size_values := ARRAY['35', '36', '37', '38', '39', '40', '41', '42', '43', '44', '45', '46', '47', '48'];
        size_units := ARRAY['EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU', 'EU'];
    END IF;
    
    -- Create array of pre-joined size+unit pairs
    size_count := array_length(size_values, 1);
    size_pairs := ARRAY[]::TEXT[];
    FOR i IN 1..size_count LOOP
        size_pairs := array_append(size_pairs, size_values[i] || size_units[i]);
    END LOOP;
    
    -- For each product we want to insert
    FOR product_id IN 1..1000 LOOP
        -- Generate random sizes for this product (3-5 sizes)
        size_list := '';
        num_sizes := floor(random()*3 + 3)::int; -- 3-5 sizes
        
        FOR i IN 1..num_sizes LOOP
            IF i > 1 THEN
                size_list := size_list || ',';
            END IF;
            -- Add random size+unit pair
            size_list := size_list || size_pairs[floor(random()*size_count + 1)::int];
        END LOOP;
        
        -- Insert product with random sizes
        INSERT INTO master_products (
            artikel, warna, size, grup, unit, kat, model, gender, 
            tipe, harga, tanggal_produk, tanggal_terima, usia, 
            status, supplier, diupdate_oleh, tanggal_update
        ) VALUES (
            'ART-' || LPAD(CAST(product_id AS TEXT), 6, '0'),
            color_names[floor(random()*color_count + 1)],
            size_list,
            (ARRAY['Casual', 'Formal', 'Sport', 'Kids', 'Accessories'])[floor(random()*5 + 1)],
            (ARRAY['PCS', 'BOX', 'SET'])[floor(random()*3 + 1)],
            (ARRAY['Premium', 'Regular', 'Basic'])[floor(random()*3 + 1)],
            'MODEL-' || floor(random()*100 + 1),
            (ARRAY['Pria', 'Wanita', 'Unisex'])[floor(random()*3 + 1)],
            (ARRAY['Baju', 'Celana', 'Jaket', 'Topi', 'Sepatu'])[floor(random()*5 + 1)],
            floor(random()*(1000000-50000 + 1) + 50000)::numeric(15,2),
            CURRENT_DATE - (floor(random()*365)::int || ' days')::interval,
            CURRENT_DATE - (floor(random()*180)::int || ' days')::interval,
            floor(random()*365 + 1)::int,
            (ARRAY['Active', 'Inactive', 'Discontinued'])[floor(random()*3 + 1)],
            (ARRAY['Supplier A', 'Supplier B', 'Supplier C', 'Supplier D', 'Supplier E'])[floor(random()*5 + 1)],
            (ARRAY['Admin1', 'Admin2', 'Admin3', 'System'])[floor(random()*4 + 1)],
            CURRENT_TIMESTAMP
        )
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
    END LOOP;
        
    RAISE NOTICE 'Added/updated sample products in the master_products table';
END $$;