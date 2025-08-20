-- Database initialization script for Go Loyalty & Benefits Platform
-- This script creates all necessary tables and initial data

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User balances table
CREATE TABLE IF NOT EXISTS balances (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    available_points BIGINT NOT NULL DEFAULT 0,
    lifetime_points BIGINT NOT NULL DEFAULT 0,
    tier VARCHAR(50) DEFAULT 'bronze',
    tier_points BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL,
    mcc VARCHAR(10) NOT NULL,
    merchant_id VARCHAR(100) NOT NULL,
    merchant_name VARCHAR(255),
    points INTEGER NOT NULL,
    multiplier DECIMAL(3,2) NOT NULL DEFAULT 1.0,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Benefits table
CREATE TABLE IF NOT EXISTS benefits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    points INTEGER NOT NULL,
    partner VARCHAR(100) NOT NULL,
    category VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT true,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    image_url VARCHAR(500),
    terms_conditions TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Redemptions table
CREATE TABLE IF NOT EXISTS redemptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    benefit_id UUID NOT NULL REFERENCES benefits(id) ON DELETE CASCADE,
    points INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'requested',
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    partner_ref VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Outbox table for event sourcing
CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    topic VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dispatched_at TIMESTAMPTZ,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- email, sms, push
    subject VARCHAR(255),
    message TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    channel VARCHAR(20) NOT NULL, -- email, sms, push
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ,
    error TEXT
);

-- Partner configurations table
CREATE TABLE IF NOT EXISTS partner_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    soap_endpoint VARCHAR(500),
    rest_endpoint VARCHAR(500),
    username VARCHAR(100),
    password_hash VARCHAR(255),
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    retry_count INTEGER NOT NULL DEFAULT 3,
    circuit_breaker_threshold INTEGER NOT NULL DEFAULT 5,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Activity logs table (for MongoDB-like functionality in Postgres)
CREATE TABLE IF NOT EXISTS activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    activity_type VARCHAR(100) NOT NULL,
    description TEXT,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_mcc ON transactions(mcc);

CREATE INDEX IF NOT EXISTS idx_redemptions_user_id ON redemptions(user_id);
CREATE INDEX IF NOT EXISTS idx_redemptions_status ON redemptions(status);
CREATE INDEX IF NOT EXISTS idx_redemptions_created_at ON redemptions(created_at);
CREATE INDEX IF NOT EXISTS idx_redemptions_idempotency_key ON redemptions(idempotency_key);

CREATE INDEX IF NOT EXISTS idx_benefits_active ON benefits(active);
CREATE INDEX IF NOT EXISTS idx_benefits_category ON benefits(category);
CREATE INDEX IF NOT EXISTS idx_benefits_partner ON benefits(partner);

CREATE INDEX IF NOT EXISTS idx_outbox_topic ON outbox(topic);
CREATE INDEX IF NOT EXISTS idx_outbox_dispatched_at ON outbox(dispatched_at);
CREATE INDEX IF NOT EXISTS idx_outbox_retry_count ON outbox(retry_count);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_activity_type ON activity_logs(activity_type);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);

-- Create functions for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for automatic timestamp updates
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_benefits_updated_at BEFORE UPDATE ON benefits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_redemptions_updated_at BEFORE UPDATE ON redemptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_partner_configs_updated_at BEFORE UPDATE ON partner_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data
INSERT INTO users (id, email, password_hash, role, first_name, last_name) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'admin@loyalty.com', crypt('admin123', gen_salt('bf')), 'admin', 'Admin', 'User'),
    ('550e8400-e29b-41d4-a716-446655440001', 'user@example.com', crypt('user123', gen_salt('bf')), 'user', 'John', 'Doe'),
    ('550e8400-e29b-41d4-a716-446655440002', 'jane@example.com', crypt('jane123', gen_salt('bf')), 'user', 'Jane', 'Smith')
ON CONFLICT (email) DO NOTHING;

INSERT INTO balances (user_id, available_points, lifetime_points, tier) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 2500, 5000, 'silver'),
    ('550e8400-e29b-41d4-a716-446655440002', 1500, 3000, 'bronze')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO benefits (id, name, description, points, partner, category, active) VALUES
    ('660e8400-e29b-41d4-a716-446655440000', '$25 Gift Card', 'Redeemable at major retailers', 2000, 'GIFTCO', 'Retail', true),
    ('660e8400-e29b-41d4-a716-446655440001', 'Free Movie Ticket', 'Valid at participating theaters', 1500, 'ENTERTAINMENTCO', 'Entertainment', true),
    ('660e8400-e29b-41d4-a716-446655440002', '$50 Travel Credit', 'Use towards flights or hotels', 4000, 'TRAVELCO', 'Travel', true),
    ('660e8400-e29b-41d4-a716-446655440003', 'Coffee Shop Gift Card', 'Valid at popular coffee chains', 800, 'RETAILCO', 'Dining', true)
ON CONFLICT DO NOTHING;

INSERT INTO partner_configs (partner_id, name, soap_endpoint, username, password_hash, timeout_seconds, retry_count) VALUES
    ('GIFTCO', 'Gift Card Company', 'https://api.giftco.com/soap', 'loyalty_user', crypt('secret', gen_salt('bf')), 30, 3),
    ('TRAVELCO', 'Travel Company', 'https://api.travelco.com/soap', 'loyalty_user', crypt('secret', gen_salt('bf')), 45, 2),
    ('RETAILCO', 'Retail Company', 'https://api.retailco.com/soap', 'loyalty_user', crypt('secret', gen_salt('bf')), 20, 3),
    ('ENTERTAINMENTCO', 'Entertainment Company', 'https://api.entertainmentco.com/soap', 'loyalty_user', crypt('secret', gen_salt('bf')), 25, 3)
ON CONFLICT (partner_id) DO NOTHING;

-- Create a view for user summary
CREATE OR REPLACE VIEW user_summary AS
SELECT 
    u.id,
    u.email,
    u.first_name,
    u.last_name,
    u.role,
    b.available_points,
    b.lifetime_points,
    b.tier,
    COUNT(t.id) as total_transactions,
    COUNT(r.id) as total_redemptions,
    u.created_at
FROM users u
LEFT JOIN balances b ON u.id = b.user_id
LEFT JOIN transactions t ON u.id = t.user_id
LEFT JOIN redemptions r ON u.id = r.user_id
GROUP BY u.id, b.available_points, b.lifetime_points, b.tier;

-- Grant permissions (adjust as needed for your setup)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO loyalty;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO loyalty;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO loyalty;

-- Create a function to get user points summary
CREATE OR REPLACE FUNCTION get_user_points_summary(user_uuid UUID)
RETURNS TABLE(
    user_id UUID,
    email VARCHAR,
    available_points BIGINT,
    lifetime_points BIGINT,
    tier VARCHAR,
    recent_transactions BIGINT,
    recent_redemptions BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.id,
        u.email,
        b.available_points,
        b.lifetime_points,
        b.tier,
        (SELECT COUNT(*) FROM transactions t WHERE t.user_id = u.id AND t.created_at > NOW() - INTERVAL '30 days'),
        (SELECT COUNT(*) FROM redemptions r WHERE r.user_id = u.id AND r.created_at > NOW() - INTERVAL '30 days')
    FROM users u
    LEFT JOIN balances b ON u.id = b.user_id
    WHERE u.id = user_uuid;
END;
$$ LANGUAGE plpgsql;

COMMENT ON TABLE users IS 'User accounts for the loyalty system';
COMMENT ON TABLE balances IS 'User loyalty point balances';
COMMENT ON TABLE transactions IS 'Loyalty point earning transactions';
COMMENT ON TABLE benefits IS 'Available benefits and rewards';
COMMENT ON TABLE redemptions IS 'Benefit redemption requests';
COMMENT ON TABLE outbox IS 'Event outbox for reliable message delivery';
COMMENT ON TABLE notifications IS 'User notifications (email, SMS, push)';
COMMENT ON TABLE partner_configs IS 'External partner service configurations';
COMMENT ON TABLE activity_logs IS 'User activity audit trail';
