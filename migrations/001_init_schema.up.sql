-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Table: roles
CREATE TABLE roles (
    id          INTEGER         PRIMARY KEY,
    name        VARCHAR(50)     NOT NULL UNIQUE,
    description TEXT            NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now()
);

-- Table: users
CREATE TABLE users (
    id           UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    email        VARCHAR(254)    NOT NULL UNIQUE,
    password_hash TEXT           NOT NULL,
    role_id      INTEGER         NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    status       VARCHAR(20)     NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'blocked')),
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ     NULL DEFAULT now(),

    CONSTRAINT users_email_check CHECK (
        char_length(email) BETWEEN 3 AND 254
        AND email !~ '[[:space:]]'
        AND email ~ '^[^@]+@[^@]+\.[^@]+$'
    )
);

-- Table: profiles
CREATE TABLE profiles (
    user_id     UUID            PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name  VARCHAR(200)    NULL,
    last_name   VARCHAR(300)    NULL,
    age         INTEGER         NULL,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ     NULL DEFAULT now()
);

-- Table: accounts
CREATE TABLE accounts (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID            NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    balance     NUMERIC(21,2)   NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ     NULL DEFAULT now(),

    CONSTRAINT accounts_user_id_unique UNIQUE (user_id)
);

-- Table: transaction_types
CREATE TABLE transaction_types (
    id          INTEGER         PRIMARY KEY,
    name        VARCHAR(50)     NOT NULL UNIQUE
);

-- Table: transactions
CREATE TABLE transactions (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id      UUID            NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    type_id         INTEGER         NOT NULL REFERENCES transaction_types(id) ON DELETE RESTRICT,
    amount          NUMERIC(21,2)   NOT NULL CHECK (amount >= 0),
    status          VARCHAR(20)     NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    -- сумма >= 0 всегда (column CHECK); нулевая сумма разрешена только для deposit (id = 1),
    -- запрещена для withdraw / transfer_in / transfer_out (id = 2, 3, 4 — см. сид 002)
    CONSTRAINT transactions_amount_zero_check CHECK (
        NOT (amount = 0 AND type_id IN (2, 3, 4))
    )
);

-- Table: sessions
CREATE TABLE sessions (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash  BYTEA           NOT NULL,
    expires_at          TIMESTAMPTZ     NOT NULL,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_refresh_token_hash ON sessions(refresh_token_hash);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_profiles_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();