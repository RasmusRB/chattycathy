-- Initialize database with seed data
-- This script runs automatically on first container start

-- Seed pings table with dummy data (table created by GORM migrations)
-- Using DO block to handle case where table doesn't exist yet
DO $$
BEGIN
    -- Wait a moment for GORM to create the table (if API starts first)
    -- This is a fallback; the API's AutoMigrate should create the table
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'pings') THEN
        INSERT INTO pings (message, created_at)
        SELECT 'pong', NOW() - (random() * interval '30 days')
        FROM generate_series(1, 10)
        WHERE NOT EXISTS (SELECT 1 FROM pings LIMIT 1);
    END IF;
END $$;
