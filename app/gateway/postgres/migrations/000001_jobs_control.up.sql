begin;

create table if not exists jobs_control
(
    job                text        primary key,
    last_success_run   timestamptz,
    is_enabled         boolean     not null default true
);

create table if not exists users
(
    id             bigint   generated always as identity  primary key,
    name           text     not null,
    phone_number   text     not null,

    created_at              timestamptz not null default current_timestamp
);


create table if not exists user_messages
(
    id             bigint      generated always as identity  primary key,
    user_id        bigint      not null references users(id),
    message        text,

    created_at                 timestamptz not null default current_timestamp
);

commit;
