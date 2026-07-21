CREATE TABLE booking_batches (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    type VARCHAR(30) NOT NULL,

    space_id BIGINT NOT NULL,
    slot_id BIGINT NULL,

    weekday SMALLINT NULL,

    start_date DATE NOT NULL,
    end_date DATE NOT NULL,

    reason VARCHAR(150) NOT NULL,

    created_by BIGINT NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'active',

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_booking_batches_space
        FOREIGN KEY (space_id)
        REFERENCES spaces(id),

    CONSTRAINT fk_booking_batches_slot
        FOREIGN KEY (slot_id)
        REFERENCES space_slots(id),

    CONSTRAINT fk_booking_batches_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id),

    CONSTRAINT booking_batches_type_check
        CHECK (type IN ('recurring_teacher', 'maintenance')),

    CONSTRAINT booking_batches_status_check
        CHECK (status IN ('active', 'cancelled')),

    CONSTRAINT booking_batches_weekday_check
        CHECK (weekday IS NULL OR weekday BETWEEN 0 AND 6),

    CONSTRAINT booking_batches_date_range_check
        CHECK (end_date >= start_date)
);

CREATE INDEX idx_booking_batches_space
ON booking_batches (space_id);

CREATE INDEX idx_booking_batches_status
ON booking_batches (status);

ALTER TABLE bookings
    ADD COLUMN batch_id BIGINT NULL REFERENCES booking_batches(id);

CREATE INDEX idx_bookings_batch_id
ON bookings (batch_id);
