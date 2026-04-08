-- Create payments table: one per order (linked via order_id)
CREATE TABLE payments (
    id                  UUID PRIMARY KEY,
    order_id            UUID NOT NULL REFERENCES orders(id),
    stripe_session_id   VARCHAR(255) UNIQUE,  -- Stripe Checkout Session ID
    ref_number          VARCHAR(50) NOT NULL UNIQUE,
    total_amount        TEXT NOT NULL,         -- customtypes.Price serialised as string
    currency            VARCHAR(10) NOT NULL,  -- ISO 4217 code
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_stripe_session_id ON payments(stripe_session_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_ref_number ON payments(ref_number);