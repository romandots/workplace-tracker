CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS shifts (
    id SERIAL PRIMARY KEY,
    employee_id INT NOT NULL REFERENCES employees(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    ip TEXT,
    user_agent TEXT,
    auto_closed BOOLEAN DEFAULT FALSE
);
