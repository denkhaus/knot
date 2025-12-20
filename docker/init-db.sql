-- Create additional database schemas if needed
-- This script runs when the PostgreSQL container starts

-- Create extensions that might be useful
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE knot_test TO knot_user;

-- Create a simple test table to verify connection
CREATE TABLE IF NOT EXISTS connection_test (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    message TEXT DEFAULT 'Database connection successful!'
);

-- Insert a test record
INSERT INTO connection_test (message) VALUES ('Knot MCP test database initialized successfully');