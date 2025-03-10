-- Create table for storing file metadata
CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    filepath VARCHAR(255) NOT NULL UNIQUE,  -- Add UNIQUE constraint
    size BIGINT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- Create GIN index for fast key-value lookups
CREATE INDEX IF NOT EXISTS idx_metadata_gin ON files USING GIN (metadata);
