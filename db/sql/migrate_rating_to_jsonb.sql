-- Migration script to change rating field from NUMERIC to JSONB
-- This script will reset all existing ratings to random values

-- Step 1: Create a backup table (optional, but recommended)
CREATE TABLE IF NOT EXISTS master_products_rating_backup AS
SELECT no, artikel, rating FROM master_products WHERE rating IS NOT NULL;

-- Step 2: Drop the existing rating column and recreate as JSONB
ALTER TABLE master_products DROP COLUMN IF EXISTS rating;
ALTER TABLE master_products ADD COLUMN rating JSONB;

-- Step 3: Set random rating values for all existing products
-- This will generate: comfort/style/support (0-10), purpose (random from predefined list)
UPDATE master_products 
SET rating = jsonb_build_object(
    'comfort', floor(random()*11)::int,
    'style', floor(random()*11)::int,
    'support', floor(random()*11)::int,
    'purpose', jsonb_build_array(
        (ARRAY['casual', 'formal', 'sport', 'outdoor', 'work', 'party', 'daily', 'running', 'walking', 'hiking', 'business', 'travel', 'gym', 'beach', 'winter'])[floor(random()*15 + 1)]
    )
)
WHERE rating IS NULL;

-- Step 4: Verify the migration
SELECT 
    COUNT(*) as total_products,
    COUNT(rating) as products_with_rating,
    COUNT(CASE WHEN rating->>'comfort' IS NOT NULL THEN 1 END) as products_with_comfort
FROM master_products;

-- Step 5: Display sample of migrated data
-- Sample output: {"comfort": 7, "style": 4, "support": 9, "purpose": ["casual"]}
SELECT artikel, rating FROM master_products LIMIT 5;