-- Seeder for master_genders table

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS master_genders (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tanggal_hapus TIMESTAMPTZ
);

-- Create index on value field
CREATE INDEX IF NOT EXISTS idx_master_genders_value ON master_genders(value);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_genders_value'
    ) THEN
        ALTER TABLE master_genders ADD CONSTRAINT uq_master_genders_value UNIQUE (value);
    END IF;
END$$;

-- Insert gender values
INSERT INTO master_genders (value, tanggal_update) VALUES
('Pria', CURRENT_TIMESTAMP),
('Wanita', CURRENT_TIMESTAMP),
('Unisex', CURRENT_TIMESTAMP)
ON CONFLICT (value) DO UPDATE
SET tanggal_update = CURRENT_TIMESTAMP; 