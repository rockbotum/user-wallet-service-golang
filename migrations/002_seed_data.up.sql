-- Seed: roles
INSERT INTO roles (id, name, description) VALUES
    (1, 'user', 'Regular user with standard permissions'),
    (2, 'admin', 'Administrator with full access')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description;

-- Seed: transaction_types
-- 1 = deposit, 2 = withdraw, 3 = transfer_in, 4 = transfer_out
INSERT INTO transaction_types (id, name) VALUES
    (1, 'deposit'),
    (2, 'withdraw'),
    (3, 'transfer_in'),
    (4, 'transfer_out')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name;