

CREATE TABLE IF NOT EXISTS topic_follows(
    user_id INTEGER NOT NULL,
    topic_id INTEGER NOT NULL,
    followed_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    UNIQUE(user_id,topic_id)
);
