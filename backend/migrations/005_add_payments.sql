ALTER TABLE bookings
    ADD COLUMN deposit_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN deposit_status VARCHAR(20) NOT NULL DEFAULT 'unpaid'
        CHECK (deposit_status IN ('unpaid', 'pending', 'paid', 'refunded')),
    ADD COLUMN deposit_method VARCHAR(20) NULL
        CHECK (deposit_method IN ('efectivo', 'transferencia', 'posnet', 'mercado_pago')),
    ADD COLUMN deposit_recorded_by BIGINT NULL REFERENCES users(id),
    ADD COLUMN expires_at TIMESTAMP NULL,
    ADD COLUMN balance_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN balance_status VARCHAR(20) NOT NULL DEFAULT 'unpaid'
        CHECK (balance_status IN ('unpaid', 'pending', 'paid', 'refunded')),
    ADD COLUMN balance_method VARCHAR(20) NULL
        CHECK (balance_method IN ('efectivo', 'transferencia', 'posnet', 'mercado_pago')),
    ADD COLUMN balance_recorded_by BIGINT NULL REFERENCES users(id);

UPDATE bookings SET
    deposit_amount = ROUND(total_price * 0.15, 2),
    balance_amount = total_price - ROUND(total_price * 0.15, 2);

ALTER TABLE bookings
    DROP CONSTRAINT bookings_status_check;

ALTER TABLE bookings
    ADD CONSTRAINT bookings_status_check
        CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed', 'expired'));

CREATE INDEX idx_bookings_pending_expiry
ON bookings (status, expires_at)
WHERE status = 'pending';

CREATE TABLE payments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    booking_id BIGINT NOT NULL REFERENCES bookings(id),
    kind VARCHAR(20) NOT NULL DEFAULT 'deposit' CHECK (kind IN ('deposit', 'balance')),
    provider VARCHAR(30) NOT NULL DEFAULT 'mercadopago',
    preference_id VARCHAR(120),
    external_payment_id VARCHAR(120),
    status VARCHAR(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'in_process', 'refunded')),
    amount DECIMAL(10, 2) NOT NULL,
    raw_status_detail VARCHAR(120),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_payments_external_payment_id UNIQUE (external_payment_id)
);

CREATE INDEX idx_payments_booking_id
ON payments (booking_id);

CREATE TRIGGER set_updated_at_payments
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
