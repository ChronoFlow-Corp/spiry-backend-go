create table if not exists users (
    id uuid primary key,
    email char(80) unique,
    access_token_google char(400),
    refresh_token_google char(400),
    refresh_token char(400)
)