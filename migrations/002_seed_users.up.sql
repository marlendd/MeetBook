INSERT INTO users (id, email, role) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@example.com', 'admin'),
    ('00000000-0000-0000-0000-000000000002', 'user@example.com', 'user')
ON CONFLICT (id) DO NOTHING;
