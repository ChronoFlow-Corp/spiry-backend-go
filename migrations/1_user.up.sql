CREATE TABLE IF NOT EXISTS users
(
    id                   uuid PRIMARY KEY,
    email                varchar(256) unique not null,
    avatar_url           varchar(1024)       not null,
    name                 varchar(256)        not null,
    last_name            varchar(256),
    theme                varchar(16)         not null default 'auto',
    google_access_token  varchar(512)        not null,
    google_refresh_token varchar(512),
    created_at           timestamptz                  default current_timestamp,
    updated_at           timestamptz                  default current_timestamp
);

CREATE TABLE IF NOT EXISTS subscriptions
(
    id               uuid PRIMARY KEY,
    name             varchar(64) not null,
    modalities_quote jsonb,
    period           varchar(32) not null,
    price            decimal(10, 2),
    level            int         not null,
    created_at       timestamptz default current_timestamp,
    updated_at       timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS plans
(
    id               uuid PRIMARY KEY,
    modalities_quote jsonb,
    user_id          uuid references users (id) on delete cascade not null,
    subscription_id  uuid references subscriptions (id)           not null,
    created_at       timestamptz default current_timestamp,
    updated_at       timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS models
(
    id         uuid PRIMARY KEY,
    name       varchar(128) not null,
    min_level  int          not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS tools
(
    id         uuid PRIMARY KEY,
    name       varchar(64)   not null,
    modalities varchar(64)[] not null,
    settings   jsonb         not null,
    prompt     text,
    min_level  int           not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS sessions
(
    id         uuid PRIMARY KEY,
    token      text         not null,
    expires_at timestamptz  not null,
    last_login timestamptz  not null,
    device     varchar(128) not null,
    user_id    uuid references users (id) on delete cascade,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);


CREATE TABLE IF NOT EXISTS chats
(
    id         uuid PRIMARY KEY,
    title      varchar(256) not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp,
    user_id    uuid references users (id) on delete cascade
);

CREATE TABLE IF NOT EXISTS commands
(
    id         uuid PRIMARY KEY,
    prompt     text                                         not null,
    settings   jsonb,
    flags      varchar(64)[],
    chat_id    uuid references chats (id) on delete cascade not null,
    tool_id    uuid references tools (id),
    model_id   uuid references models (id),
    status     varchar(64)                                  not null,
    user_id    uuid references users (id),
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS command_medias
(
    id         uuid PRIMARY KEY,
    name       varchar(128)                                    not null,
    type       varchar(256)                                    not null,
    url        varchar(256)                                    not null unique,
    size       int                                             not null,
    command_id uuid references commands (id) on delete cascade not null,
    user_id    uuid references users (id) on delete cascade,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS results
(
    id             uuid PRIMARY KEY,
    open_router_id varchar(64)                                     not null,
    Text           text                                            not null,
    command_id     uuid references commands (id) on delete cascade not null,
    tool_id        uuid references tools (id)                      not null,
    chat_id        uuid references chats (id) on delete cascade    not null,
    user_id        uuid references users (id) on delete cascade,
    model_id       uuid references models (id)                     not null,
    created_at     timestamptz default current_timestamp,
    updated_at     timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS result_medias
(
    id         uuid PRIMARY KEY,
    name       varchar(128)                                   not null,
    type       varchar(256)                                   not null,
    url        varchar(256)                                   not null,
    size       int                                            not null,
    result_id  uuid references results (id) on delete cascade not null,
    user_id    uuid references users (id) on delete cascade,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

CREATE TABLE IF NOT EXISTS admins
(
    id         uuid PRIMARY KEY,
    login      varchar(64)  not null,
    password   varchar(128) not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);