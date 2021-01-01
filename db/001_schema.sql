CREATE TABLE IF NOT EXISTS users
(
    id       BIGSERIAL PRIMARY KEY, /* it is a serial just for simplicity */
    email    VARCHAR(320) NOT NULL UNIQUE,
    password VARCHAR(200) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx on users (email);

CREATE TABLE IF NOT EXISTS groups
(
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users_groups
(
    user_id  BIGINT NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS expenses
(
    id        BIGSERIAL PRIMARY KEY,
    user_id   BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount    REAL      NOT NULL,
    /* its recommended not to call columns as reserved words, did so in accordance with the description */
    timestamp TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE INDEX IF NOT EXISTS expenses_user_id_idx on expenses (user_id);

CREATE TABLE IF NOT EXISTS expenses_shares
(
    expense_id BIGINT   NOT NULL REFERENCES expenses (id) ON DELETE CASCADE,
    user_id    BIGINT   NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    percent    SMALLINT NOT NULL
);

CREATE INDEX IF NOT EXISTS expenses_shares_user_id_idx on expenses_shares (user_id);
