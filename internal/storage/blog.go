package storage

import (
	"encoding/json"
	"errors"
	"time"
)

type BlogStatus string

// 'draft','published','archived'
const (
	BlogStatusDraft     BlogStatus = "draft"
	BlogStatusPublished BlogStatus = "published"
	BlogStatusArchived  BlogStatus = "archived"
)

type Blog struct {
	Id              int             `db:"id" json:"id"`
	BlogTitle       string          `db:"blog_title" json:"blog_title"`
	BlogDescription *string         `db:"blog_description" json:"blog_description"`
	BlogContent     json.RawMessage `db:"blog_content" json:"blog_content"`
	BlogThumbnail   *string         `db:"blog_thumbnail" json:"blog_thumbnail"`
	BlogStatus      BlogStatus      `db:"blog_status" json:"blog_status"`
	BlogAuthorId    int             `db:"blog_author_id" json:"blog_author_id"`
	PublishedAt     *string         `db:"published_at" json:"published_at"`
	BlogCreatedAt   string          `db:"blog_created_at" json:"blog_created_at"`
	BlogUpdatedAt   *string         `db:"blog_updated_at" json:"blog_updated_at"`
}

type BlogTopic struct {
	BlogId  int `db:"blog_id" json:"blog_id"`
	TopicId int `db:"topic_id" json:"topic_id"`
}

type BlogWithUserAndTopics struct {
	Blog
	BlogAuthor User    `json:"blog_author"`
	BlogTopics []Topic `json:"blog_topics"`
}

func (s *Storage) CreateBlogWithTopics(blogTitle string, blogDescription string, blogContent json.RawMessage, blogThumbnail string, blogStatus BlogStatus, blogAuthorId int, topicIds []int) (*BlogWithUserAndTopics, error) {

	var createdBlogPost BlogWithUserAndTopics

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	var rollBackErr error
	defer func() {
		if rollBackErr != nil {
			tx.Rollback()
		}
	}()

	var blog Blog
	insertBlogQuery := `INSERT INTO blogs(blog_title,blog_description,blog_content,blog_thumbnail,blog_status,blog_author_id,published_at) 
	VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,blog_title,blog_description,blog_content,blog_thumbnail,blog_status,
	blog_author_id,published_at,blog_created_at,blog_updated_at`

	var publishedAtArg any
	if blogStatus == BlogStatusPublished {
		publishedAtArg = time.Now()
	} else {
		publishedAtArg = nil
	}

	if err := tx.QueryRowx(insertBlogQuery, blogTitle, blogDescription, blogContent, blogThumbnail, blogStatus, blogAuthorId, publishedAtArg).StructScan(&blog); err != nil {
		rollBackErr = err
		return nil, rollBackErr
	}

	var topics []Topic
	for _, topicId := range topicIds {
		var blogTopic BlogTopic
		var topic Topic

		blogTopicQuery := `INSERT INTO blog_topics(blog_id,topic_id) VALUES($1,$2) RETURNING blog_id,topic_id`
		if err := tx.QueryRowx(blogTopicQuery, blog.Id, topicId).StructScan(&blogTopic); err != nil {
			rollBackErr = err
			return nil, rollBackErr
		}

		topicQuery := `SELECT id,topic_name,created_at,updated_at 
		FROM topics WHERE id=$1`

		if err := tx.QueryRowx(topicQuery, blogTopic.TopicId).StructScan(&topic); err != nil {
			rollBackErr = err
			return nil, rollBackErr
		}

		topics = append(topics, topic)
	}

	var blogAuthor User
	blogAuthorQuery := `SELECT id,email,username,password,name,profile_img,
    is_verified,role,created_at,updated_at FROM users WHERE id=$1`

	if err := tx.QueryRowx(blogAuthorQuery, blog.BlogAuthorId).StructScan(&blogAuthor); err != nil {
		rollBackErr = err
		return nil, rollBackErr
	}

	createdBlogPost.Blog = blog
	createdBlogPost.BlogTopics = topics
	createdBlogPost.BlogAuthor = blogAuthor

	if rollBackErr = tx.Commit(); rollBackErr != nil {
		return nil, rollBackErr
	}

	return &createdBlogPost, nil
}

func (s *Storage) CreateBlog(blogTitle string, blogDescription string, blogContent json.RawMessage, blogThumbnail string, blogStatus BlogStatus, blogAuthorId int) (*BlogWithUserAndTopics, error) {
	var createdBlog BlogWithUserAndTopics

	var blog Blog
	insertBlogQuery := `INSERT INTO blogs(blog_title,blog_description,blog_content,blog_thumbnail,blog_status,blog_author_id,published_at) 
	VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,blog_title,blog_description,blog_content,blog_thumbnail,blog_status,
	blog_author_id,published_at,blog_created_at,blog_updated_at`

	var publishedAtArg any
	if blogStatus == BlogStatusPublished {
		publishedAtArg = time.Now()
	} else {
		publishedAtArg = nil
	}

	if err := s.db.QueryRowx(insertBlogQuery, blogTitle, blogDescription, blogContent, blogThumbnail, blogStatus, blogAuthorId, publishedAtArg).StructScan(&blog); err != nil {
		return nil, err
	}

	var blogAuthor User
	blogAuthorQuery := `SELECT id, email, username, password, name, profile_img, is_verified, role, created_at, updated_at 
	FROM users WHERE id=$1`

	if err := s.db.QueryRowx(blogAuthorQuery, blog.BlogAuthorId).StructScan(&blogAuthor); err != nil {
		return nil, err
	}

	createdBlog.Blog = blog
	createdBlog.BlogAuthor = blogAuthor
	return &createdBlog, nil
}

func (s *Storage) GetBlogById(blogId int) (*Blog, error) {

	var blog Blog

	query := `SELECT id,id, blog_title, blog_description, blog_content, blog_thumbnail, blog_status, blog_author_id, published_at, blog_created_at, blog_updated_at 
	FROM blogs WHERE id=$1`

	if err := s.db.QueryRowx(query, blogId).StructScan(&blog); err != nil {
		return nil, err
	}

	return &blog, nil
}

func (s *Storage) DeleteBlogById(blogId int) error {

	query := `DELETE FROM blogs WHERE id=$1`

	result, err := s.db.Exec(query, blogId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("blog not found")
	}

	return nil
}

func (s *Storage) GetBlogTopics(blogId int) ([]Topic, error) {
	var topics []Topic

	query := `SELECT id, topic_name, created_at, updated_at 
	FROM topics WHERE id IN (SELECT topic_id FROM blog_topics WHERE blog_id=$1)`

	rows, err := s.db.Queryx(query, blogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic Topic

		if err := rows.StructScan(&topic); err != nil {
			return nil, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

// update blog status from 'draft' to 'published' and add additional topics to blog
func (s *Storage) PublishBlogAndAddTopics(blogId int, topicIds []int) (*BlogWithUserAndTopics, error) {

	var updatedBlog BlogWithUserAndTopics

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	var rollBackErr error
	defer func() {
		if rollBackErr != nil {
			tx.Rollback()
		}
	}()

	var blog Blog
	// update blog status to 'published' query
	updateBlogStatusQuery := `UPDATE blogs SET blog_status=$1,published_at=$2 WHERE id=$3 
RETURNING id,blog_title,blog_description,blog_content,blog_thumbnail,blog_status,blog_author_id,published_at,
blog_created_at,blog_updated_at`
	//	add topics to blog

	if err := tx.QueryRowx(updateBlogStatusQuery, BlogStatusPublished, time.Now(), blogId).StructScan(&blog); err != nil {
		rollBackErr = err
		return nil, rollBackErr
	}

	insertTopicQuery := `INSERT INTO blog_topics(blog_id,topic_id) VALUES($1,$2)`

	// these are additional topicIds (not necessarily all blog topicIds)
	for _, topicId := range topicIds {
		_, err := tx.Exec(insertTopicQuery, blog.Id, topicId)
		if err != nil {
			rollBackErr = err
			return nil, rollBackErr
		}
	}

	var topics []Topic
	// get blog topics now (all topics , existing + added)
	blogTopicsQuery := `SELECT id,topic_name,created_at,updated_at 
	FROM topics WHERE id IN (SELECT topic_id FROM blog_topics WHERE blog_id=$1)`

	topicRows, err := tx.Queryx(blogTopicsQuery, blog.Id)
	if err != nil {
		rollBackErr = err
		return nil, rollBackErr
	}
	defer topicRows.Close()

	for topicRows.Next() {
		var topic Topic

		if err := topicRows.StructScan(&topic); err != nil {
			rollBackErr = err
			return nil, rollBackErr
		}

		topics = append(topics, topic)
	}

	var blogAuthor User
	blogAuthorQuery := `SELECT id,email,username,password,name,profile_img,
    is_verified,role,created_at,updated_at FROM users WHERE id=$1`

	if err := tx.QueryRowx(blogAuthorQuery, blog.BlogAuthorId).StructScan(&blogAuthor); err != nil {
		rollBackErr = err
		return nil, rollBackErr
	}

	if err = tx.Commit(); err != nil {
		rollBackErr = err
		return nil, rollBackErr
	}

	updatedBlog.Blog = blog
	updatedBlog.BlogTopics = topics
	updatedBlog.BlogAuthor = blogAuthor

	return &updatedBlog, nil
}

func (s *Storage) UpdateBlogStatus(blogId int, blogStatus BlogStatus) (*BlogWithUserAndTopics, error) {

	var updatedBlog BlogWithUserAndTopics

	var blog Blog
	query := `UPDATE blogs SET blog_status=$1,published_at=$2 WHERE id=$3 
	RETURNING id,blog_title,blog_description,blog_content,blog_thumbnail,blog_status,blog_author_id,published_at,
	blog_created_at,blog_updated_at`

	var publishedAtArg any
	if blogStatus == BlogStatusPublished {
		publishedAtArg = time.Now()
	} else {
		publishedAtArg = nil
	}

	if err := s.db.QueryRowx(query, blogStatus, publishedAtArg, blogId).StructScan(&blog); err != nil {
		return nil, err
	}

	var topics []Topic

	topicsQuery := `SELECT id,topic_name,created_at,updated_at 
	FROM topics WHERE id IN (SELECT topic_id FROM blog_topics WHERE blog_id=$1)`
	topicRows, err := s.db.Queryx(topicsQuery, blog.Id)
	if err != nil {
		return nil, err
	}
	defer topicRows.Close()

	for topicRows.Next() {
		var topic Topic

		if err := topicRows.StructScan(&topic); err != nil {
			return nil, err
		}

		topics = append(topics, topic)
	}

	var blogAuthor User
	blogAuthorQuery := `SELECT id, email, username, password, name, profile_img, is_verified, role, created_at, updated_at 
	FROM users WHERE id=$1`

	if err := s.db.QueryRowx(blogAuthorQuery, blog.BlogAuthorId).StructScan(&blogAuthor); err != nil {
		return nil, err
	}

	updatedBlog.Blog = blog
	updatedBlog.BlogTopics = topics
	updatedBlog.BlogAuthor = blogAuthor

	return &updatedBlog, nil
}
