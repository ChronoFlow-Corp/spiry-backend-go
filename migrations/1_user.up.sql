create table if not exists users (
    id uuid primary key,
    email varchar(80) unique NOT NULL,
    name varchar(64) NOT NULL,
    picture_url varchar(256) not null,
    language varchar(16) default 'en',
    admin bool default false NOT NULL,
    theme varchar(16) default 'system' NOT NULL,
    access_token_google text,
    refresh_token_google text,
    refresh_token  text,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
)
