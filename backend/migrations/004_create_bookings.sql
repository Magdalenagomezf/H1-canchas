CREATE TABLE bookings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    customer_user_id BIGINT,
    created_by BIGINT NOT NULL,

    customer_name VARCHAR(100),
    customer_phone VARCHAR(30),

    space_id BIGINT NOT NULL,
    slot_id BIGINT NOT NULL,

    booking_date DATE NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'pending',

    total_price DECIMAL(10, 2) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_bookings_customer_user
        FOREIGN KEY (customer_user_id)
        REFERENCES users(id),

    CONSTRAINT fk_bookings_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id),

    CONSTRAINT fk_bookings_space
        FOREIGN KEY (space_id)
        REFERENCES spaces(id),

    CONSTRAINT fk_bookings_slot
        FOREIGN KEY (slot_id)
        REFERENCES space_slots(id),

    CONSTRAINT bookings_status_check
        CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),

    CONSTRAINT bookings_customer_info_check
        CHECK (
            customer_user_id IS NOT NULL
            OR
            (
                customer_name IS NOT NULL
                AND customer_phone IS NOT NULL
            )
        )
);

CREATE UNIQUE INDEX unique_active_booking
ON bookings (space_id, slot_id, booking_date)
WHERE status IN ('pending', 'confirmed');

CREATE INDEX idx_bookings_date
ON bookings (booking_date);

CREATE INDEX idx_bookings_space_date
ON bookings (space_id, booking_date);

CREATE INDEX idx_bookings_customer_user
ON bookings (customer_user_id);

CREATE INDEX idx_bookings_created_by
ON bookings (created_by);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at_bookings
BEFORE UPDATE ON bookings
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();