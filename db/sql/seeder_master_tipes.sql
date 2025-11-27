-- Seeder for master_tipes table

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS master_tipes (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tanggal_hapus TIMESTAMPTZ
);

-- Create index on value field
CREATE INDEX IF NOT EXISTS idx_master_tipes_value ON master_tipes(value);

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_tipes_value'
    ) THEN
        ALTER TABLE master_tipes ADD CONSTRAINT uq_master_tipes_value UNIQUE (value);
    END IF;
END$$;

-- Insert tipe values
INSERT INTO master_tipes (value, tanggal_update) VALUES
('Baju', CURRENT_TIMESTAMP),
('Celana', CURRENT_TIMESTAMP),
('Jaket', CURRENT_TIMESTAMP),
('Topi', CURRENT_TIMESTAMP),
('Sepatu', CURRENT_TIMESTAMP)
ON CONFLICT (value) DO UPDATE
SET tanggal_update = CURRENT_TIMESTAMP; 

-- Remove unique constraint after insertions
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_tipes_value'
    ) THEN
        ALTER TABLE master_tipes DROP CONSTRAINT uq_master_tipes_value;
    END IF;
END$$;