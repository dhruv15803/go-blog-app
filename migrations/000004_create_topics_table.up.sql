



CREATE TABLE IF NOT EXISTS topics(
    id SERIAL PRIMARY KEY,
    topic_name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);
