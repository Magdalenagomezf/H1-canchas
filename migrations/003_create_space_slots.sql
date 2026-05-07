CREATE TABLE space_slots (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    space_id BIGINT NOT NULL,

    label VARCHAR(100) NOT NULL,

    description TEXT,

    start_time TIME,
    end_time TIME,



    is_active BOOLEAN NOT NULL DEFAULT true,

    CONSTRAINT fk_space_slots_space
        FOREIGN KEY (space_id)
        REFERENCES spaces(id),

    CONSTRAINT space_slots_time_check
        CHECK (
            (start_time IS NULL AND end_time IS NULL)
            OR
            (start_time IS NOT NULL AND end_time IS NOT NULL AND end_time > start_time)
        ),

    CONSTRAINT space_slots_unique_label
        UNIQUE (space_id, label)
);