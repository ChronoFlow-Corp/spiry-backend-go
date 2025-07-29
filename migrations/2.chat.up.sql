create table if not exists chats (
     id uuid primary key,
     title char(80),
     user_id uuid references users(id)
);

create table if not exists messages (
    id uuid primary key,
    question text,
    answer text,
    chat_id uuid references chats(id),
    user_id uuid references users(id)
);

