-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    stripe_payment_method_id VARCHAR(255) UNIQUE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('card', 'bank_account', 'sepa_debit')),
    is_default BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    card_brand VARCHAR(50),
    card_last4 VARCHAR(4),
    card_exp_month INTEGER,
    card_exp_year INTEGER,
    fingerprint VARCHAR(255),
    bank_name VARCHAR(255),
    bank_last4 VARCHAR(4),
    sepa_debit_country VARCHAR(2),
    sepa_debit_bank_code VARCHAR(8),
    sepa_debit_last4 VARCHAR(4),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraint
ALTER TABLE payment_methods ADD CONSTRAINT fk_payment_methods_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE INDEX idx_payment_methods_user_id ON payment_methods(user_id);
CREATE INDEX idx_payment_methods_stripe_payment_method_id ON payment_methods(stripe_payment_method_id);
CREATE INDEX idx_payment_methods_type ON payment_methods(type);
CREATE INDEX idx_payment_methods_is_default ON payment_methods(is_default);
CREATE INDEX idx_payment_methods_fingerprint ON payment_methods(fingerprint);

-- Add comments for documentation
COMMENT ON TABLE payment_methods IS 'User payment methods with Stripe integration';
COMMENT ON COLUMN payment_methods.user_id IS 'User who owns this payment method';
COMMENT ON COLUMN payment_methods.stripe_payment_method_id IS 'Stripe payment method ID';
COMMENT ON COLUMN payment_methods.type IS 'Payment method type: card, bank_account, sepa_debit';
COMMENT ON COLUMN payment_methods.is_default IS 'Whether this is the default payment method';
COMMENT ON COLUMN payment_methods.metadata IS 'Additional metadata in JSON format';
COMMENT ON COLUMN payment_methods.card_brand IS 'Card brand (visa, mastercard, etc.)';
COMMENT ON COLUMN payment_methods.card_last4 IS 'Last 4 digits of card';
COMMENT ON COLUMN payment_methods.card_exp_month IS 'Card expiration month';
COMMENT ON COLUMN payment_methods.card_exp_year IS 'Card expiration year';
COMMENT ON COLUMN payment_methods.fingerprint IS 'Card fingerprint for uniqueness';
COMMENT ON COLUMN payment_methods.bank_name IS 'Bank name for bank accounts';
COMMENT ON COLUMN payment_methods.bank_last4 IS 'Last 4 digits of bank account';
COMMENT ON COLUMN payment_methods.sepa_debit_country IS 'SEPA debit country code';
COMMENT ON COLUMN payment_methods.sepa_debit_bank_code IS 'SEPA debit bank code';
COMMENT ON COLUMN payment_methods.sepa_debit_last4 IS 'Last 4 digits of SEPA debit account';