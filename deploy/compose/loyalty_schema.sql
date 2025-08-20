-- Loyalty Service Database Schema
-- This script creates the necessary tables for the loyalty service

-- Create loyalty_users table
CREATE TABLE IF NOT EXISTS loyalty_users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    points INTEGER DEFAULT 0 NOT NULL,
    tier VARCHAR(50) DEFAULT 'Bronze' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create loyalty_transactions table
CREATE TABLE IF NOT EXISTS loyalty_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('earn', 'spend')),
    amount INTEGER NOT NULL CHECK (amount > 0),
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES loyalty_users(id) ON DELETE CASCADE
);

-- Create loyalty_rewards table
CREATE TABLE IF NOT EXISTS loyalty_rewards (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    points_cost INTEGER NOT NULL CHECK (points_cost > 0),
    category VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_loyalty_users_email ON loyalty_users(email);
CREATE INDEX IF NOT EXISTS idx_loyalty_users_tier ON loyalty_users(tier);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_user_id ON loyalty_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_created_at ON loyalty_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_loyalty_rewards_category ON loyalty_rewards(category);
CREATE INDEX IF NOT EXISTS idx_loyalty_rewards_points_cost ON loyalty_rewards(points_cost);
CREATE INDEX IF NOT EXISTS idx_loyalty_rewards_active ON loyalty_rewards(is_active);

-- Insert sample rewards
INSERT INTO loyalty_rewards (id, name, description, points_cost, category, is_active) VALUES
    ('reward-001', 'Free Coffee', 'Redeem for a free coffee at any participating location', 100, 'Food & Beverage', true),
    ('reward-002', 'Movie Ticket', 'Redeem for a movie ticket at any participating theater', 500, 'Entertainment', true),
    ('reward-003', 'Amazon Gift Card', '$10 Amazon gift card', 1000, 'Shopping', true),
    ('reward-004', 'Restaurant Discount', '20% off at participating restaurants', 200, 'Food & Beverage', true),
    ('reward-005', 'Gas Station Credit', '$5 credit at participating gas stations', 250, 'Transportation', true),
    ('reward-006', 'Hotel Upgrade', 'Room upgrade at participating hotels', 2000, 'Travel', true),
    ('reward-007', 'Free Shipping', 'Free shipping on your next order', 150, 'Shopping', true),
    ('reward-008', 'Concert Ticket', 'Redeem for a concert ticket', 1500, 'Entertainment', true)
ON CONFLICT (id) DO NOTHING;

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_loyalty_users_updated_at 
    BEFORE UPDATE ON loyalty_users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_loyalty_rewards_updated_at 
    BEFORE UPDATE ON loyalty_rewards 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to calculate tier based on points
CREATE OR REPLACE FUNCTION calculate_tier(points INTEGER)
RETURNS VARCHAR(50) AS $$
BEGIN
    IF points >= 10000 THEN
        RETURN 'Platinum';
    ELSIF points >= 5000 THEN
        RETURN 'Gold';
    ELSIF points >= 1000 THEN
        RETURN 'Silver';
    ELSE
        RETURN 'Bronze';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create function to update user tier automatically
CREATE OR REPLACE FUNCTION update_user_tier()
RETURNS TRIGGER AS $$
BEGIN
    NEW.tier = calculate_tier(NEW.points);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update tier when points change
CREATE TRIGGER update_loyalty_users_tier 
    BEFORE UPDATE ON loyalty_users 
    FOR EACH ROW EXECUTE FUNCTION update_user_tier();

-- Insert sample loyalty users (for testing)
INSERT INTO loyalty_users (id, email, points, tier) VALUES
    ('user-001', 'john.doe@example.com', 1500, 'Silver'),
    ('user-002', 'jane.smith@example.com', 500, 'Bronze'),
    ('user-003', 'bob.wilson@example.com', 7500, 'Gold')
ON CONFLICT (id) DO NOTHING;

-- Insert sample transactions (for testing)
INSERT INTO loyalty_transactions (id, user_id, type, amount, description) VALUES
    ('tx-001', 'user-001', 'earn', 100, 'Purchase at Coffee Shop'),
    ('tx-002', 'user-001', 'earn', 200, 'Restaurant dining'),
    ('tx-003', 'user-001', 'spend', 100, 'Redeemed free coffee'),
    ('tx-004', 'user-002', 'earn', 150, 'Gas station purchase'),
    ('tx-005', 'user-003', 'earn', 500, 'Hotel stay'),
    ('tx-006', 'user-003', 'earn', 300, 'Shopping at department store')
ON CONFLICT (id) DO NOTHING;

-- Grant permissions (adjust as needed for your setup)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO loyalty;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO loyalty;


