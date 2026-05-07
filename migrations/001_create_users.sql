CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,

    name VARCHAR(100) NOT NULL,

    email VARCHAR(150) UNIQUE,
    phone VARCHAR(30),

    password_hash TEXT,

    role VARCHAR(30) NOT NULL DEFAULT 'customer',

    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
