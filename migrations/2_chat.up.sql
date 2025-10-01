create table if not exists chats (
     id uuid primary key,
     title char(80) NOT NULL,
     user_id uuid references users(id) ON DELETE CASCADE NOT NULL,
     created_at timestamp default current_timestamp,
     updated_at timestamp default current_timestamp
);

create table if not exists messages (
    id uuid primary key,
    content text,
    role char(32) NOT NULL,
    chat_id uuid references chats(id) ON DELETE CASCADE NOT NULL,
    user_id uuid references users(id) ON DELETE CASCADE NOT NULL,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

