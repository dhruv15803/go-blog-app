package storage

import (
	"errors"
	"time"
)

type BlogComment struct {
	Id                 int     `db:"id" json:"id"`
	BlogCommentContent string  `db:"blog_comment" json:"blog_comment"`
	CommentAuthorId    int     `db:"comment_author_id" json:"comment_author_id"`
	BlogId             int     `db:"blog_id" json:"blog_id"`
	ParentCommentId    *int    `db:"parent_comment_id" json:"parent_comment_id"`
	CommentCreatedAt   string  `db:"comment_created_at" json:"comment_created_at"`
	CommentUpdatedAt   *string `db:"comment_updated_at" json:"comment_updated_at"`
}

type BlogCommentWithAuthor struct {
	BlogComment
	BlogCommentAuthor User `json:"blog_comment_author"`
}

type BlogCommentWithMetaData struct {
	BlogComment
	BlogCommentAuthor        User `json:"blog_comment_author"`
	BlogCommentLikesCount    int  `json:"blog_comment_likes_count"`
	BlogCommentCommentsCount int  `json:"blog_comment_comments_count"`
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

// CreateBlogComment creating a top level blog comment
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

func (s *Storage) UpdateBlogCommentById(id int, blogCommentContent string) (*BlogCommentWithAuthor, error) {

	var blogCommentWithAuthor BlogCommentWithAuthor

	var blogComment BlogComment
	query := `UPDATE blog_comments SET blog_comment=$1,comment_updated_at=$2 WHERE id=$3 RETURNING 
	id,blog_comment,comment_author_id,blog_id,parent_comment_id,comment_created_at,comment_updated_at`

	if err := s.db.QueryRowx(query, blogCommentContent, time.Now(), id).StructScan(&blogComment); err != nil {
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

// GetBlogComments gets top level comments for blog (parent_comment_id==null)
func (s *Storage) GetBlogComments(blogId int, skip int, limit int) ([]BlogCommentWithMetaData, error) {

	var blogComments []BlogCommentWithMetaData

	query := `SELECT
  bc.id,
  bc.blog_comment,
  bc.comment_author_id,
  bc.blog_id,
  bc.parent_comment_id,
  bc.comment_created_at,
  bc.comment_updated_at,
  u.id,
  u.email,
  u.username,
  u.password,
  u.name,
  u.profile_img,
  u.is_verified,
  u.role,
  u.created_at,
  u.updated_at,
  COUNT(DISTINCT cl.liked_by_id) AS blog_comment_likes_count,
  COUNT(DISTINCT cc.id) AS blog_comment_comments_count
FROM
  blog_comments AS bc
  INNER JOIN users AS u ON bc.comment_author_id = u.id
  LEFT JOIN blog_comment_likes AS cl ON bc.id = cl.liked_blog_comment_id
  LEFT JOIN blog_comments AS cc ON bc.id = cc.parent_comment_id
WHERE
  bc.blog_id = $1 AND bc.parent_comment_id IS NULL
GROUP BY
  bc.id,u.id
ORDER BY
  bc.comment_created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := s.db.Queryx(query, blogId, limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var blogComment BlogCommentWithMetaData

		if err := rows.Scan(&blogComment.Id, &blogComment.BlogCommentContent, &blogComment.CommentAuthorId,
			&blogComment.BlogId, &blogComment.ParentCommentId, &blogComment.CommentCreatedAt, &blogComment.CommentUpdatedAt,
			&blogComment.BlogCommentAuthor.Id, &blogComment.BlogCommentAuthor.Email, &blogComment.BlogCommentAuthor.Username,
			&blogComment.BlogCommentAuthor.Password, &blogComment.BlogCommentAuthor.Name, &blogComment.BlogCommentAuthor.ProfileImg,
			&blogComment.BlogCommentAuthor.IsVerified, &blogComment.BlogCommentAuthor.Role, &blogComment.BlogCommentAuthor.CreatedAt, &blogComment.BlogCommentAuthor.UpdatedAt,
			&blogComment.BlogCommentLikesCount, &blogComment.BlogCommentCommentsCount); err != nil {
			return nil, err
		}

		blogComments = append(blogComments, blogComment)
	}

	return blogComments, nil
}

// top level comments count for blog
func (s *Storage) GetBlogCommentsCount(blogId int) (int, error) {
	var totalBlogCommentsCount int

	query := `SELECT COUNT(id) FROM blog_comments WHERE blog_id=$1 AND parent_comment_id IS NULL`

	if err := s.db.QueryRowx(query, blogId).Scan(&totalBlogCommentsCount); err != nil {
		return -1, err
	}

	return totalBlogCommentsCount, nil
}

// get child comments for a blog comment (parent_comment_id=blogCommentId)
func (s *Storage) GetChildBlogComments(blogCommentId int, skip int, limit int) ([]BlogCommentWithMetaData, error) {

	var blogComments []BlogCommentWithMetaData

	query := `SELECT
  bc.id,
  bc.blog_comment,
  bc.comment_author_id,
  bc.blog_id,
  bc.parent_comment_id,
  bc.comment_created_at,
  bc.comment_updated_at,
  u.id,
  u.email,
  u.username,
  u.password,
  u.name,
  u.profile_img,
  u.is_verified,
  u.role,
  u.created_at,
  u.updated_at,
  COUNT(DISTINCT cl.liked_by_id) AS blog_comment_likes_count,
  COUNT(DISTINCT cc.id) AS blog_comment_comments_count
FROM
  blog_comments AS bc
  INNER JOIN users AS u ON bc.comment_author_id = u.id
  LEFT JOIN blog_comment_likes AS cl ON bc.id = cl.liked_blog_comment_id
  LEFT JOIN blog_comments AS cc ON bc.id = cc.parent_comment_id
WHERE
  bc.parent_comment_id=$1
GROUP BY
  bc.id,u.id
ORDER BY
  bc.comment_created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := s.db.Queryx(query, blogCommentId, limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var blogComment BlogCommentWithMetaData

		if err := rows.Scan(&blogComment.Id, &blogComment.BlogCommentContent, &blogComment.CommentAuthorId,
			&blogComment.BlogId, &blogComment.ParentCommentId, &blogComment.CommentCreatedAt, &blogComment.CommentUpdatedAt,
			&blogComment.BlogCommentAuthor.Id, &blogComment.BlogCommentAuthor.Email, &blogComment.BlogCommentAuthor.Username,
			&blogComment.BlogCommentAuthor.Password, &blogComment.BlogCommentAuthor.Name, &blogComment.BlogCommentAuthor.ProfileImg,
			&blogComment.BlogCommentAuthor.IsVerified, &blogComment.BlogCommentAuthor.Role, &blogComment.BlogCommentAuthor.CreatedAt, &blogComment.BlogCommentAuthor.UpdatedAt,
			&blogComment.BlogCommentLikesCount, &blogComment.BlogCommentCommentsCount); err != nil {
			return nil, err
		}

		blogComments = append(blogComments, blogComment)
	}

	return blogComments, nil
}

func (s *Storage) GetChildBlogCommentsCount(blogCommentId int) (int, error) {
	var totalBlogCommentsCount int

	query := `SELECT COUNT(id) FROM blog_comments WHERE parent_comment_id=$1`

	if err := s.db.QueryRowx(query, blogCommentId).Scan(&totalBlogCommentsCount); err != nil {
		return -1, err
	}

	return totalBlogCommentsCount, nil
}
