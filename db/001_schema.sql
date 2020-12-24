create table if not exists users
(
    id bigint primary key,
    email varchar(500) not null unique,
    password varchar(200) not null
);
