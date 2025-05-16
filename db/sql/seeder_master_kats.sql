-- Seeder for master_kats table

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS master_kats (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tanggal_hapus TIMESTAMPTZ
);

-- Create index on value field
CREATE INDEX IF NOT EXISTS idx_master_kats_value ON master_kats(value);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_kats_value'
    ) THEN
        ALTER TABLE master_kats ADD CONSTRAINT uq_master_kats_value UNIQUE (value);
    END IF;
END$$;

-- Insert category values
INSERT INTO master_kats (value, tanggal_update) VALUES
('Premium', CURRENT_TIMESTAMP),
('Regular', CURRENT_TIMESTAMP),
('Basic', CURRENT_TIMESTAMP)
ON CONFLICT (value) DO UPDATE
SET tanggal_update = CURRENT_TIMESTAMP; 