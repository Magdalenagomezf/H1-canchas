-- CI only: the admin the E2E tests use to create and clean their data.
-- Password: e2e-admin-password (the CI database is thrown away after each run).
INSERT INTO users (name, phone, email, password_hash, role)
VALUES ('E2E Admin', '+5493515000001', 'e2e-admin@example.com',
        '$2a$10$7ZkjnCv4ZRoL0N3ugF/lKuxCAdIcrbYzfLgAZeQkEboXDa4nEcw8G', 'admin');
