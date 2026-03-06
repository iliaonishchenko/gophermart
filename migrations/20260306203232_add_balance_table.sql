-- +goose Up
CREATE TABLE balances(
    user_id UUID PRIMARY KEY REFERENCES users(id),
    balance NUMERIC
);

-- +goose Down
DROP TABLE balances;
