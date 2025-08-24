package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
)

type CreateBlogCommentRequest struct {
	BlogComment     string `json:"blog_comment"`
	BlogId          int    `json:"blog_id"`
	ParentCommentId *int   `json:"parent_comment_id"`
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
