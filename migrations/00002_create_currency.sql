-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS currencies (
    id BIGSERIAL PRIMARY KEY UNIQUE NOT NULL,
    code CHAR(3) UNIQUE NOT NULL
);

INSERT INTO currencies(code)
VALUES('RUB'),
      ('USD'),
      ('EUR'),
      ('CNY');

CREATE TABLE IF NOT EXISTS user_wallets (
    id BIGSERIAL PRIMARY KEY UNIQUE NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    currency_id BIGINT NOT NULL REFERENCES currencies(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT unique_user_currency UNIQUE (user_id, currency_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_wallets CASCADE;
DROP TABLE currencies CASCADE;
-- +goose StatementEnd
