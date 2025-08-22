

CREATE TYPE user_role AS ENUM('user','admin');

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE,
    password TEXT NOT NULL,
    name TEXT,
    profile_img TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    role user_role DEFAULT ('user'),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);
