BEGIN;

ALTER TABLE spaces
    DROP CONSTRAINT IF EXISTS spaces_type_check;

ALTER TABLE spaces
    ADD CONSTRAINT spaces_type_check
        CHECK (type IN ('cancha_padel', 'cancha_futbol', 'cancha_padbol', 'cancha_beach_voley', 'quincho'));

COMMIT;
