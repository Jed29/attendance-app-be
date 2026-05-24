CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE offices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    radius_meters INT NOT NULL DEFAULT 100,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE employees (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    office_id     UUID NOT NULL REFERENCES offices(id),
    name          VARCHAR(255) NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50) NOT NULL DEFAULT 'employee',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE attendances (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id  UUID NOT NULL REFERENCES employees(id),
    check_in_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    check_out_at TIMESTAMPTZ,
    check_in_lat DOUBLE PRECISION NOT NULL,
    check_in_lng DOUBLE PRECISION NOT NULL,
    photo_url    VARCHAR(500),
    status       VARCHAR(50) NOT NULL DEFAULT 'present',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE leave_requests (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id  UUID NOT NULL REFERENCES employees(id),
    leave_type   VARCHAR(50) NOT NULL,
    start_date   DATE NOT NULL,
    end_date     DATE NOT NULL,
    reason       TEXT,
    status       VARCHAR(50) NOT NULL DEFAULT 'pending',
    approved_by  UUID REFERENCES employees(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE shifts (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id            UUID NOT NULL REFERENCES employees(id),
    start_time             TIME NOT NULL,
    end_time               TIME NOT NULL,
    late_tolerance_minutes INT NOT NULL DEFAULT 15
);

CREATE INDEX idx_attendances_employee_date ON attendances(employee_id, check_in_at);
CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id);
