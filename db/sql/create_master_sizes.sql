CREATE TABLE
    IF NOT EXISTS master_sizes (
        id SERIAL PRIMARY KEY,
        value TEXT NOT NULL UNIQUE,
        unit TEXT NOT NULL,
        tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        tanggal_hapus TIMESTAMPTZ
    );

-- Create index on value for faster searches
CREATE INDEX IF NOT EXISTS idx_master_sizes_value ON master_sizes(value);

-- Remove the old unique constraint on value and unit combination
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_master_sizes_value_unit'
    ) THEN
        ALTER TABLE master_sizes DROP CONSTRAINT uq_master_sizes_value_unit;
    END IF;
END$$; 