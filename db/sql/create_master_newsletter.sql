-- SQL for creating the master_newsletter table
CREATE TABLE IF NOT EXISTS master_newsletter (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    whatsapp TEXT NOT NULL,
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_master_newsletter_email ON master_newsletter(email);
