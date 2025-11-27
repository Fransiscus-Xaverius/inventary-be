-- Seeder for master_grups table

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS master_grups (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tanggal_hapus TIMESTAMPTZ
);

-- Create index on value field
CREATE INDEX IF NOT EXISTS idx_master_grups_value ON master_grups(value);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_grups_value'
    ) THEN
        ALTER TABLE master_grups ADD CONSTRAINT uq_master_grups_value UNIQUE (value);
    END IF;
END$$;

-- Insert grup values
INSERT INTO master_grups (value, tanggal_update) VALUES
('Casual', CURRENT_TIMESTAMP),
('Formal', CURRENT_TIMESTAMP),
('Sport', CURRENT_TIMESTAMP),
('Kids', CURRENT_TIMESTAMP),
('Accessories', CURRENT_TIMESTAMP)
ON CONFLICT (value) DO UPDATE
SET tanggal_update = CURRENT_TIMESTAMP; 

-- Remove unique constraint after insertions
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_grups_value'
    ) THEN
        ALTER TABLE master_grups DROP CONSTRAINT uq_master_grups_value;
    END IF;
END$$;