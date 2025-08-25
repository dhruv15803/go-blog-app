package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type CreateBlogCommentRequest struct {
	BlogComment     string `json:"blog_comment"`
	BlogId          int    `json:"blog_id"`
	ParentCommentId *int   `json:"parent_comment_id"`
}

type UpdateBlogCommentRequest struct {
	BlogComment string `json:"blog_comment"`
}

func (h *Handler) CreateBlogCommentHandler(w http.ResponseWriter, r *http.Request) {

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user does not exist", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	var createBlogCommentPayload CreateBlogCommentRequest

	if err := json.NewDecoder(r.Body).Decode(&createBlogCommentPayload); err != nil {
		writeJSONError(w, "invalid request", http.StatusBadRequest)
		return
	}

	blogComment := strings.TrimSpace(createBlogCommentPayload.BlogComment)
	blogId := createBlogCommentPayload.BlogId
	parentCommentId := createBlogCommentPayload.ParentCommentId
	isTopLevelComment := parentCommentId == nil

	if blogComment == "" {
		writeJSONError(w, "blog comment content cannot be empty", http.StatusBadRequest)
		return
	}

	blog, err := h.storage.GetBlogById(blogId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if isTopLevelComment {
		//	parent comment id is null, so this is a top level comment for the blog
		//	 create a top level comment (no child comment)
		blogComment, err := h.storage.CreateBlogComment(blogComment, user.Id, blog.Id)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success     bool                          `json:"success"`
			Message     string                        `json:"message"`
			BlogComment storage.BlogCommentWithAuthor `json:"blog_comment"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "created blog comment", BlogComment: *blogComment}, http.StatusCreated); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	} else {

		//	parent commend id is not null , this is a child comment for the blog
		//	parent comment id has to exist and be comment of this blog
		parentComment, err := h.storage.GetBlogCommentById(*parentCommentId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, "parent comment not found", http.StatusBadRequest)
				return
			} else {
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		if parentComment.BlogId != blog.Id {
			writeJSONError(w, "parent comment is not blog's comment", http.StatusBadRequest)
			return
		}

		//	parentComment.BlogId == blog.Id

		// create blog with parent_comment_id=parentComment.Id (child comment)
		childBlogComment, err := h.storage.CreateChildBlogComment(blogComment, user.Id, blog.Id, parentComment.Id)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success     bool                          `json:"success"`
			Message     string                        `json:"message"`
			BlogComment storage.BlogCommentWithAuthor `json:"blog_comment"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "created child blog comment", BlogComment: *childBlogComment}, http.StatusCreated); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) LikeBlogCommentHandler(w http.ResponseWriter, r *http.Request) {

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user does not exist", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	blogCommentId, err := strconv.ParseInt(chi.URLParam(r, "blogCommentId"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid request param blogCommentId", http.StatusBadRequest)
		return
	}

	blogComment, err := h.storage.GetBlogCommentById(int(blogCommentId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog comment not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	blogCommentLike, err := h.storage.GetBlogCommentLike(user.Id, blogComment.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if blogCommentLike == nil {

		blogCommentLike, err := h.storage.CreateBlogCommentLike(user.Id, blogComment.Id)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success         bool                    `json:"success"`
			Message         string                  `json:"message"`
			BlogCommentLike storage.BlogCommentLike `json:"blog_comment_like"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "liked blog comment", BlogCommentLike: *blogCommentLike}, http.StatusCreated); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	} else {

		if err := h.storage.RemoveBlogCommentLike(blogCommentLike.LikedById, blogCommentLike.LikedBlogCommentId); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "removed blog comment like"}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) DeleteBlogCommentHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user does not exist", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	blogCommentId, err := strconv.ParseInt(chi.URLParam(r, "blogCommentId"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid request param blogCommentId", http.StatusBadRequest)
		return
	}

	blogComment, err := h.storage.GetBlogCommentById(int(blogCommentId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog comment not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if user.Id != blogComment.CommentAuthorId {
		writeJSONError(w, "unauthorized to delete blog comment", http.StatusUnauthorized)
		return
	}

	if err := h.storage.DeleteBlogCommentById(blogComment.Id); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "deleted blog comment"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateBlogCommentHandler(w http.ResponseWriter, r *http.Request) {

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user does not exist", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	blogCommentId, err := strconv.ParseInt(chi.URLParam(r, "blogCommentId"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid request param blogCommentId", http.StatusBadRequest)
		return
	}

	blogComment, err := h.storage.GetBlogCommentById(int(blogCommentId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog comment not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if user.Id != blogComment.CommentAuthorId {
		writeJSONError(w, "unauthorized to update blog comment", http.StatusUnauthorized)
		return
	}

	var updateBlogCommentPayload UpdateBlogCommentRequest

	if err := json.NewDecoder(r.Body).Decode(&updateBlogCommentPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newBlogCommentContent := strings.TrimSpace(updateBlogCommentPayload.BlogComment)

	if newBlogCommentContent == "" {
		writeJSONError(w, "blog comment content is required", http.StatusBadRequest)
		return
	}

	updatedBlogComment, err := h.storage.UpdateBlogCommentById(blogComment.Id, newBlogCommentContent)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success     bool                          `json:"success"`
		Message     string                        `json:"message"`
		BlogComment storage.BlogCommentWithAuthor `json:"blog_comment"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "updated blog comment", BlogComment: *updatedBlogComment}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) GetBlogCommentsHandler(w http.ResponseWriter, r *http.Request) {

	blogId, err := strconv.ParseInt(chi.URLParam(r, "blogId"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid request param blogId", http.StatusBadRequest)
		return
	}

	blog, err := h.storage.GetBlogById(int(blogId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog does not exist", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
	var page int
	var limit int

	if r.URL.Query().Get("page") == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			writeJSONError(w, "invalid query param page", http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Get("limit") == "" {
		limit = 10
	} else {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			writeJSONError(w, "invalid query param limit", http.StatusBadRequest)
			return
		}
	}

	skip := page*limit - limit

	//	get blog comments where (blog_id = blog.Id) order by created-at limit 10 offset 0

	blogComments, err := h.storage.GetBlogComments(blog.Id, skip, limit)
	if err != nil {
		log.Printf("failed to get blog comments: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	totalBlogCommentsCount, err := h.storage.GetBlogCommentsCount(blog.Id)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	noOfPages := int(math.Ceil(float64(totalBlogCommentsCount) / float64(limit)))

	type Response struct {
		Success      bool                              `json:"success"`
		BlogComments []storage.BlogCommentWithMetaData `json:"blog_comments"`
		NoOfPages    int                               `json:"no_of_pages"`
	}

	if err := writeJSON(w, Response{Success: true, BlogComments: blogComments, NoOfPages: noOfPages}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

// get comments  for a blog comment (child comments)
func (h *Handler) GetBlogCommentCommentsHandler(w http.ResponseWriter, r *http.Request) {

	blogCommentId, err := strconv.ParseInt(chi.URLParam(r, "blogCommentId"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid request param blogCommentId", http.StatusBadRequest)
		return
	}

	blogComment, err := h.storage.GetBlogCommentById(int(blogCommentId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog comment does not exist", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
	var page int
	var limit int

	if r.URL.Query().Get("page") == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			writeJSONError(w, "invalid query param page", http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Get("limit") == "" {
		limit = 10
	} else {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			writeJSONError(w, "invalid query param limit", http.StatusBadRequest)
			return
		}
	}

	skip := page*limit - limit

	//	get blog comments where parent_comment_id=blogComment.Id
	blogComments, err := h.storage.GetChildBlogComments(blogComment.Id, skip, limit)
	if err != nil {
		log.Printf("failed to get blog comments: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	totalBlogsCount, err := h.storage.GetChildBlogCommentsCount(blogComment.Id)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	noOfPages := int(math.Ceil(float64(totalBlogsCount) / float64(limit)))

	type Response struct {
		Success      bool                              `json:"success"`
		BlogComments []storage.BlogCommentWithMetaData `json:"blog_comments"`
		NoOfPages    int                               `json:"no_of_pages"`
	}

	if err := writeJSON(w, Response{Success: true, BlogComments: blogComments, NoOfPages: noOfPages}, http.StatusOK); err != nil {
		writeJSON(w, "internal server error", http.StatusInternalServerError)
	}
}
