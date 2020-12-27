create table if not exists users
(
    id       SERIAL PRIMARY KEY,
    email    varchar(320) NOT NULL UNIQUE,
    password varchar(200) NOT NULL
);
