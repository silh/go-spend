create table if not exists users
(
    id       BIGSERIAL PRIMARY KEY,
    email    varchar(320) NOT NULL UNIQUE,
    password varchar(200) NOT NULL
);

create table if not exists groups
(
    id   BIGSERIAL PRIMARY KEY,
    name varchar(100) NOT NULL UNIQUE
);

create table if not exists users_groups
(
    user_id  BIGINT NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups (id) ON DELETE CASCADE
);
