


CREATE TABLE IF NOT EXISTS blog_comments(
    id SERIAL PRIMARY KEY,
    blog_comment TEXT NOT NULL,
    comment_author_id INTEGER NOT NULL,
    blog_id INTEGER NOT NULL,
    parent_comment_id INTEGER,
    comment_created_at TIMESTAMP DEFAULT NOW(),
    comment_updated_at TIMESTAMP,
    FOREIGN KEY(comment_author_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(blog_id) REFERENCES  blogs(id) ON DELETE CASCADE,
    FOREIGN KEY(parent_comment_id) REFERENCES blog_comments(id) ON DELETE CASCADE
);
