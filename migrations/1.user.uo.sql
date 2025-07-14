create table if not exists users (
    id uuid primary key ,
    email char(120),
    access_token_google char(80),
    refresh_token_google char(80),
    refresh_token char(80)
)