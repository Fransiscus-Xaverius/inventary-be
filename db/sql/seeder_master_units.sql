-- Seeder for master_units table

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS master_units (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tanggal_hapus TIMESTAMPTZ
);

-- Create index on value field
CREATE INDEX IF NOT EXISTS idx_master_units_value ON master_units(value);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_units_value'
    ) THEN
        ALTER TABLE master_units ADD CONSTRAINT uq_master_units_value UNIQUE (value);
    END IF;
END$$;

-- Insert unit values
INSERT INTO master_units (value, tanggal_update) VALUES
('PCS', CURRENT_TIMESTAMP),
('BOX', CURRENT_TIMESTAMP),
('SET', CURRENT_TIMESTAMP)
ON CONFLICT (value) DO UPDATE
SET tanggal_update = CURRENT_TIMESTAMP; 