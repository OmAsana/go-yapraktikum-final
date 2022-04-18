-- +goose Up
CREATE TABLE if not exists public.users
(
    user_id       INT GENERATED ALWAYS AS IDENTITY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    password      VARCHAR(50)        NOT NULL,
    created_at    TIMESTAMP          NOT NULL,
    last_login_at TIMESTAMP,
    PRIMARY KEY (user_id)
);

CREATE TABLE if not exists public.orders
(
    order_id     INT         NOT NULL,
    status       VARCHAR(50) NOT NULL,
    tx_type      VARCHAR(50) NOT NULL,
    accrual      NUMERIC DEFAULT 0,
    user_id      INT,
    uploaded_at  TIMESTAMP   NOT NULL,
    processed_at TIMESTAMP   NOT NULL,
    PRIMARY KEY (order_id),
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES users (user_id)
);




-- +goose Down
DROP TABLE if exists public.orders;
DROP TABLE if exists public.users;
