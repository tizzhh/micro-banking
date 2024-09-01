-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, email, pass_hash, first_name, last_name, balance, age)
VALUES (10000, 'test@gmail.com', '$2a$10$IF6t0BJ/uEZfNnKkuMExjOg/mTxq1xn.y3X7stLCCLl54nTiN5A1.', 'admin', 'admin', 1000, 20);

INSERT INTO user_wallets (user_id, currency_id, balance)
VALUES (10000, 1, 0),
       (10000, 2, 0),
       (10000, 3, 0),
       (10000, 4, 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM user_wallets WHERE user_id = 10000;
DELETE FROM users WHERE id = 10000;
-- +goose StatementEnd
