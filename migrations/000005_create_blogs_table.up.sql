
CREATE TYPE blog_status AS ENUM('draft','published','archived');

CREATE TABLE IF NOT EXISTS blogs(
    id SERIAL PRIMARY KEY,
    blog_title TEXT NOT NULL,
    blog_description TEXT,
    blog_content JSONB NOT NULL,
    blog_thumbnail TEXT,
    blog_status blog_status DEFAULT('draft'),
    blog_author_id INTEGER NOT NULL,
    published_at TIMESTAMP,
    blog_created_at TIMESTAMP DEFAULT NOW(),
    blog_updated_at TIMESTAMP,
    FOREIGN KEY(blog_author_id) REFERENCES users(id) ON DELETE CASCADE
);
