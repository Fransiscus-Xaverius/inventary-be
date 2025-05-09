INSERT INTO master_products (artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh)
SELECT 
    'ART-' || LPAD(CAST(generate_series AS TEXT), 6, '0') as artikel,
    (ARRAY['Merah', 'Biru', 'Hitam', 'Putih', 'Abu-abu', 'Hijau', 'Kuning'])[floor(random()*7 + 1)] as warna,
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
    (ARRAY['Admin1', 'Admin2', 'Admin3', 'System'])[floor(random()*4 + 1)] as diupdate_oleh
FROM generate_series(1, 1000);