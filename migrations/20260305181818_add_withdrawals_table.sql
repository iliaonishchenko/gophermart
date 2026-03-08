-- +goose Up
CREATE TABLE withdrawals(
    order_number VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    sum NUMERIC,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE withdrawals;
