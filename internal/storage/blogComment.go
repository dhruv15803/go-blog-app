package storage

import "errors"

type BlogComment struct {
	Id               int     `db:"id" json:"id"`
	BlogComment      string  `db:"blog_comment" json:"blog_comment"`
	CommentAuthorId  int     `db:"comment_author_id" json:"comment_author_id"`
	BlogId           int     `db:"blog_id" json:"blog_id"`
	ParentCommentId  *int    `db:"parent_comment_id" json:"parent_comment_id"`
	CommentCreatedAt string  `db:"comment_created_at" json:"comment_created_at"`
	CommentUpdatedAt *string `db:"comment_updated_at" json:"comment_updated_at"`
}

type BlogCommentWithAuthor struct {
	BlogComment
	BlogCommentAuthor User `json:"blog_comment_author"`
}

func (s *Storage) GetBlogCommentById(id int) (*BlogComment, error) {

	var blogComment BlogComment

	query := `SELECT id, blog_comment, comment_author_id, blog_id, parent_comment_id, comment_created_at, comment_updated_at 
	FROM blog_comments WHERE id=$1`

	if err := s.db.QueryRowx(query, id).StructScan(&blogComment); err != nil {
		return nil, err
	}

	return &blogComment, nil
}

// creating a top level blog comment
func (s *Storage) CreateBlogComment(blogCommentContent string, commentAuthorId int, blogId int) (*BlogCommentWithAuthor, error) {

	var blogCommentWithAuthor BlogCommentWithAuthor

	var blogComment BlogComment
	query := `INSERT INTO blog_comments(blog_comment,comment_author_id,blog_id) VALUES($1,$2,$3) 
	RETURNING id,blog_comment,comment_author_id,blog_id,parent_comment_id,comment_created_at,comment_updated_at`

	if err := s.db.QueryRowx(query, blogCommentContent, commentAuthorId, blogId).StructScan(&blogComment); err != nil {
		return nil, err
	}

	var blogCommentAuthor User
	commentAuthorQuery := `SELECT id, email, username, password, name, profile_img, is_verified, role, created_at, updated_at 
	FROM users WHERE id=$1`

	if err := s.db.QueryRowx(commentAuthorQuery, blogComment.CommentAuthorId).StructScan(&blogCommentAuthor); err != nil {
		return nil, err
	}

	blogCommentWithAuthor.BlogComment = blogComment
	blogCommentWithAuthor.BlogCommentAuthor = blogCommentAuthor

	return &blogCommentWithAuthor, nil
}

func (s *Storage) CreateChildBlogComment(blogCommentContent string, commentAuthorId int, blogId int, parentCommentId int) (*BlogCommentWithAuthor, error) {

	var blogCommentWithAuthor BlogCommentWithAuthor

	var blogComment BlogComment
	query := `INSERT INTO blog_comments(blog_comment,comment_author_id,blog_id,parent_comment_id) VALUES($1,$2,$3,$4) 
	RETURNING id,blog_comment,comment_author_id,blog_id,parent_comment_id,comment_created_at,comment_updated_at`

	if err := s.db.QueryRowx(query, blogCommentContent, commentAuthorId, blogId, parentCommentId).StructScan(&blogComment); err != nil {
		return nil, err
	}

	var blogCommentAuthor User
	commentAuthorQuery := `SELECT id, email, username, password, name, profile_img, is_verified, role, created_at, updated_at 
	FROM users WHERE id=$1`

	if err := s.db.QueryRowx(commentAuthorQuery, blogComment.CommentAuthorId).StructScan(&blogCommentAuthor); err != nil {
		return nil, err
	}

	blogCommentWithAuthor.BlogComment = blogComment
	blogCommentWithAuthor.BlogCommentAuthor = blogCommentAuthor
	return &blogCommentWithAuthor, nil
}

func (s *Storage) DeleteBlogCommentById(id int) error {

	query := `DELETE FROM blog_comments WHERE id=$1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("blog comment not deleted")
	}

	return nil

}
