package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type CreateBlogRequest struct {
	BlogTitle        string             `json:"blog_title"`
	BlogDescription  string             `json:"blog_description"`
	BlogContent      json.RawMessage    `json:"blog_content"`
	BlogThumbnailUrl string             `json:"blog_thumbnail_url"`
	BlogStatus       storage.BlogStatus `json:"blog_status"`
	BlogTopicIds     []int              `json:"blog_topic_ids"`
}

func (h *Handler) CreateBlogHandler(w http.ResponseWriter, r *http.Request) {

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

	var createBlogPayload CreateBlogRequest

	if err := json.NewDecoder(r.Body).Decode(&createBlogPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	blogTitle := strings.TrimSpace(createBlogPayload.BlogTitle)
	blogDescription := strings.TrimSpace(createBlogPayload.BlogDescription)
	blogContentJson := createBlogPayload.BlogContent
	blogThumbnailUrl := createBlogPayload.BlogThumbnailUrl
	blogStatus := createBlogPayload.BlogStatus
	blogTopicIds := createBlogPayload.BlogTopicIds // array of topic id's [1,4,6,7] that the blog will have

	if blogTitle == "" || len(blogContentJson) == 0 {
		writeJSONError(w, "blog title and content are required", http.StatusBadRequest)
		return
	}

	for _, topicId := range blogTopicIds {
		_, err := h.storage.GetTopicById(topicId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Printf("failed to find topic with id %d", topicId)
				writeJSONError(w, "topic does not exist", http.StatusBadRequest)
				return
			} else {
				log.Printf("failed to get topic by id: %v\n", err)
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}
	}

	if len(blogTopicIds) != 0 {

		newBlog, err := h.storage.CreateBlogWithTopics(blogTitle, blogDescription, blogContentJson, blogThumbnailUrl, blogStatus, user.Id, blogTopicIds)
		if err != nil {
			log.Printf("failed to create blog with topics: %v\n", err)
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool                          `json:"success"`
			Message string                        `json:"message"`
			Blog    storage.BlogWithUserAndTopics `json:"blog"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "created blog successfully", Blog: *newBlog}, http.StatusCreated); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	} else {
		//	no topicIds[]int array
		//	so if no topics mentioned and blogStatus = 'published' then throw error
		if blogStatus == storage.BlogStatusPublished {
			writeJSONError(w, "blog topics compulsory for published blog", http.StatusBadRequest)
			return
		}

		newBlog, err := h.storage.CreateBlog(blogTitle, blogDescription, blogContentJson, blogThumbnailUrl, blogStatus, user.Id)
		if err != nil {
			log.Printf("failed to create blog: %v\n", err)
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool                          `json:"success"`
			Message string                        `json:"message"`
			Blog    storage.BlogWithUserAndTopics `json:"blog"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "blog created successfully", Blog: *newBlog}, http.StatusCreated); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) DeleteBlogHandler(w http.ResponseWriter, r *http.Request) {

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

	if user.Id != blog.BlogAuthorId {
		writeJSONError(w, "unauthorized to delete blog", http.StatusUnauthorized)
		return
	}

	if err := h.storage.DeleteBlogById(blog.Id); err != nil {
		log.Printf("failed to delete blog: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "blog deleted successfully"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}
