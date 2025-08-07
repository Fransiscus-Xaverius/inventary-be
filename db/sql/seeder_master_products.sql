-- Seeder for master_products table
-- This seeder will add 1000 sample products

-- Create the table if it doesn't exist
DROP TABLE IF EXISTS master_products;
CREATE TABLE IF NOT EXISTS master_products (
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
    harga NUMERIC(15,2),
    harga_diskon NUMERIC(15, 2),
    marketplace JSONB,
    offline JSONB,
    gambar TEXT[],
    tanggal_produk DATE,
    tanggal_terima DATE,
    status TEXT,
    supplier TEXT,
    diupdate_oleh TEXT,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tanggal_hapus TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_master_products_artikel ON master_products(artikel);
CREATE INDEX IF NOT EXISTS idx_master_products_grup ON master_products(grup);
CREATE INDEX IF NOT EXISTS idx_master_products_offline ON master_products USING GIN (offline);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_products_artikel'
    ) THEN
        ALTER TABLE master_products ADD CONSTRAINT uq_master_products_artikel UNIQUE (artikel);
    END IF;
END$$;

-- Set up the seeder
DO $$
DECLARE
    color_ids INTEGER[];
    color_count INTEGER;
    size_list TEXT := '';
    color_list TEXT := '';
    i INTEGER;
    product_id INTEGER;
    num_sizes INTEGER;
    num_colors INTEGER;
    eu_size INTEGER;
    is_range BOOLEAN;
    
    -- Arrays to hold values from master tables
    grup_values TEXT[];
    unit_values TEXT[];
    kat_values TEXT[];
    gender_values TEXT[];
    tipe_values TEXT[];
    
    -- Count of values in each array
    grup_count INTEGER;
    unit_count INTEGER;
    kat_count INTEGER;
    gender_count INTEGER;
    tipe_count INTEGER;
    
    -- Selected random values
    random_grup TEXT;
    random_unit TEXT;
    random_kat TEXT;
    random_gender TEXT;
    random_tipe TEXT;

    nama_list TEXT[];
    deskripsi_list TEXT[];
    nama_count INTEGER;
    deskripsi_count INTEGER;
    random_nama TEXT;
    random_deskripsi TEXT;

    marketplace JSONB;
BEGIN
    -- Get all color IDs from master_colors table
    SELECT ARRAY_AGG(id) INTO color_ids FROM master_colors WHERE tanggal_hapus IS NULL;
    
    -- If no colors found, use default message
    IF color_ids IS NULL OR array_length(color_ids, 1) IS NULL THEN
        RAISE EXCEPTION 'No colors found in master_colors table';
    END IF;
    
    -- Get count of colors for random selection
    color_count := array_length(color_ids, 1);
    
    -- Get values from master tables
    SELECT ARRAY_AGG(value) INTO grup_values FROM master_grups WHERE tanggal_hapus IS NULL;
    SELECT ARRAY_AGG(value) INTO unit_values FROM master_units WHERE tanggal_hapus IS NULL;
    SELECT ARRAY_AGG(value) INTO kat_values FROM master_kats WHERE tanggal_hapus IS NULL;
    SELECT ARRAY_AGG(value) INTO gender_values FROM master_genders WHERE tanggal_hapus IS NULL;
    SELECT ARRAY_AGG(value) INTO tipe_values FROM master_tipes WHERE tanggal_hapus IS NULL;
    
    -- Get counts for random selection
    grup_count := array_length(grup_values, 1);
    unit_count := array_length(unit_values, 1);
    kat_count := array_length(kat_values, 1);
    gender_count := array_length(gender_values, 1);
    tipe_count := array_length(tipe_values, 1);
    
    -- Default values in case tables are empty
    IF grup_values IS NULL OR grup_count IS NULL THEN
        grup_values := ARRAY['Casual', 'Formal', 'Sport', 'Kids', 'Accessories'];
        grup_count := 5;
    END IF;
    
    IF unit_values IS NULL OR unit_count IS NULL THEN
        unit_values := ARRAY['PCS', 'BOX', 'SET'];
        unit_count := 3;
    END IF;
    
    IF kat_values IS NULL OR kat_count IS NULL THEN
        kat_values := ARRAY['Premium', 'Regular', 'Basic'];
        kat_count := 3;
    END IF;
    
    IF gender_values IS NULL OR gender_count IS NULL THEN
        gender_values := ARRAY['Pria', 'Wanita', 'Unisex'];
        gender_count := 3;
    END IF;
    
    IF tipe_values IS NULL OR tipe_count IS NULL THEN
        tipe_values := ARRAY['Baju', 'Celana', 'Jaket', 'Topi', 'Sepatu'];
        tipe_count := 5;
    END IF;

    -- For each product we want to insert
    FOR product_id IN 1..1000 LOOP
        -- Generate random sizes for this product (1-4 sizes)
        size_list := '';
        num_sizes := floor(random()*4 + 1)::int; -- 1-4 sizes
        
        FOR i IN 1..num_sizes LOOP
            IF i > 1 THEN
                size_list := size_list || ',';
            END IF;
            
            -- Generate EU size between 30 and 48
            eu_size := floor(random()*(48-30+1) + 30)::int;
            
            -- Decide if this will be a single size or a range (20% chance of range)
            is_range := random() < 0.2;
            
            IF is_range THEN
                -- For a range, add 1-3 to the base size
                size_list := size_list || eu_size || '-' || (eu_size + floor(random()*3 + 1)::int);
            ELSE
                -- Single size
                size_list := size_list || eu_size;
            END IF;
        END LOOP;
        
        -- Generate random colors for this product (1-3 colors)
        color_list := '';
        num_colors := floor(random()*3 + 1)::int; -- 1-3 colors
        
        FOR i IN 1..num_colors LOOP
            IF i > 1 THEN
                color_list := color_list || ',';
            END IF;
            -- Add random color ID
            color_list := color_list || color_ids[floor(random()*color_count + 1)]::text;
        END LOOP;
        
        -- Select random values from master tables
        random_grup := grup_values[floor(random()*grup_count + 1)];
        random_unit := unit_values[floor(random()*unit_count + 1)];
        random_kat := kat_values[floor(random()*kat_count + 1)];
        random_gender := gender_values[floor(random()*gender_count + 1)];
        random_tipe := tipe_values[floor(random()*tipe_count + 1)];

        -- Define nama and deskripsi list
        nama_list := ARRAY['Jogging Shoes','Running Shoes','Walking Shoes','Casual Shoes','Formal Shoes','Sports Shoes','Sandals','Flip Flops'];
        deskripsi_list := ARRAY['Lightweight and comfortable for everyday use','Perfect for running and jogging','Great for walking and casual wear','Stylish and elegant for formal occasions','Versatile for sports and outdoor activities','Comfortable for sports and fitness','Lightweight and comfortable for everyday use','Perfect for running and jogging'];

        -- Select random nama and deskripsi
        nama_count := array_length(nama_list, 1);
        deskripsi_count := array_length(deskripsi_list, 1);
        random_nama := nama_list[floor(random()*nama_count + 1)];
        random_deskripsi := deskripsi_list[floor(random()*deskripsi_count + 1)];

        -- Generate random marketplace values
        marketplace := json_build_object(
            'tokopedia', 'https://www.tokopedia.com/everysoft/product-id',
            'shopee', 'https://shopee.co.id/everysoft/product-id',
            'lazada', 'https://www.lazada.co.id/everysoft/product-id',
            'tiktok', 'https://www.tiktok.com/everysoft/product-id',
            'bukalapak', 'https://www.bukalapak.com/everysoft/product-id'
        );

        -- Insert product with random values
        INSERT INTO master_products (
            artikel, nama, deskripsi, rating, warna, size, grup, unit, kat, model, gender, 
            tipe, harga, harga_diskon, marketplace, gambar, tanggal_produk, tanggal_terima, 
            status, supplier, diupdate_oleh, tanggal_update
        ) VALUES (
            'ART-' || LPAD(CAST(product_id AS TEXT), 6, '0'),
            random_nama,
            random_deskripsi,
            jsonb_build_object(
                'comfort', floor(random()*11)::int,
                'style', floor(random()*11)::int,
                'support', floor(random()*11)::int,
                'purpose', CASE 
                    -- 70% chance of single purpose
                    WHEN random() < 0.7 THEN jsonb_build_array(
                        (ARRAY['casual', 'formal', 'sport', 'outdoor', 'work', 'party', 'daily', 'running', 'walking', 'hiking', 'business', 'travel', 'gym', 'beach', 'winter'])[floor(random()*15 + 1)]
                    )
                    -- 30% chance of multiple purposes (2-3 purposes)
                    ELSE jsonb_build_array(
                        (ARRAY['casual', 'formal', 'sport', 'outdoor', 'work', 'party', 'daily', 'running', 'walking', 'hiking', 'business', 'travel', 'gym', 'beach', 'winter'])[floor(random()*15 + 1)],
                        (ARRAY['casual', 'formal', 'sport', 'outdoor', 'work', 'party', 'daily', 'running', 'walking', 'hiking', 'business', 'travel', 'gym', 'beach', 'winter'])[floor(random()*15 + 1)]
                    )
                END
            ),
            color_list,
            size_list,
            random_grup,
            random_unit,
            random_kat,
            'MODEL-' || floor(random()*100 + 1),
            random_gender,
            random_tipe,
            floor(random()*(1000000-50000 + 1) + 50000)::numeric(15,2),
            floor(random()*(500000-40000 + 1))::numeric(15,2),
            marketplace,
            ARRAY['/uploads/products/1.png', '/uploads/products/2.png', '/uploads/products/3.png'],
            CURRENT_DATE - (floor(random()*365)::int || ' days')::interval,
            CURRENT_DATE - (floor(random()*1000)::int || ' days')::interval,
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
            status = EXCLUDED.status,
            supplier = EXCLUDED.supplier,
            diupdate_oleh = EXCLUDED.diupdate_oleh,
            tanggal_update = CURRENT_TIMESTAMP;
    END LOOP;
        
    RAISE NOTICE 'Added/updated sample products in the master_products table';
END $$;