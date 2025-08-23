package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

type UpdateBlogStatusRequest struct {
	BlogStatus   storage.BlogStatus `json:"blog_status"`
	BlogTopicIds []int              `json:"blog_topic_ids"` // optional additional topic ids that user might want to add while publishing a 'draft' blog
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

func (h *Handler) UpdateBlogStatusHandler(w http.ResponseWriter, r *http.Request) {
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
		writeJSONError(w, "unauthorized to update blog status", http.StatusUnauthorized)
		return
	}
	//	if blog's status was 'draft' and changing to 'published' , 'blog should have topics and also should have feature to add extra
	//  topics that were not added while creating blog as draft.
	var updateBlogStatusPayload UpdateBlogStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&updateBlogStatusPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newBlogStatus := updateBlogStatusPayload.BlogStatus
	additionalTopicIds := updateBlogStatusPayload.BlogTopicIds
	for _, topicId := range additionalTopicIds {
		_, err := h.storage.GetTopicById(topicId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, "topic does not exist", http.StatusBadRequest)
				return
			} else {
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}
	}

	if blog.BlogStatus == storage.BlogStatusDraft && newBlogStatus == storage.BlogStatusPublished {

		//	check if draft has any topics , if not then there better be additionalTopicIds to add to the draft and then publish
		// if there are existing topics,  then check if there are additional topics, if yes . then check that additional topics do not include any of the blog's topics
		existingBlogTopics, err := h.storage.GetBlogTopics(blog.Id)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
		var existingBlogTopicIds []int
		for _, existingBlogTopic := range existingBlogTopics {
			existingBlogTopicIds = append(existingBlogTopicIds, existingBlogTopic.Id)
		}

		var publishedBlog *storage.BlogWithUserAndTopics

		if len(existingBlogTopics) > 0 {
			//	check if any additional topic id exists in blogTopicIds
			for _, topicId := range additionalTopicIds {
				//	check if this topicId exists in existingBlogTopicIds
				if isArrayContainElement(existingBlogTopicIds, topicId) {
					writeJSONError(w, "topic already exists", http.StatusBadRequest)
					return
				}
			}

			//	if here then additional topic ids are valid and ready to be added to blog's topics
			// update blog's status from 'draft' to 'published' and 'update blog's topics' -> do in one transaction
			publishedBlog, err = h.storage.PublishBlogAndAddTopics(blog.Id, additionalTopicIds)
			if err != nil {
				log.Printf("failed to publish and add topics: %v\n", err)
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		} else {

			//	 no existing blogs so if additionTopicIds.length == 0 then throw error
			if len(additionalTopicIds) == 0 {
				writeJSONError(w, "topics required to publish blog", http.StatusBadRequest)
				return
			}

			//	so the additionalTopicIds are the all topicIds for this blog currently
			publishedBlog, err = h.storage.PublishBlogAndAddTopics(blog.Id, additionalTopicIds)
			if err != nil {
				log.Printf("failed to publish and add topics: %v\n", err)
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		type Response struct {
			Success bool                          `json:"success"`
			Message string                        `json:"message"`
			Blog    storage.BlogWithUserAndTopics `json:"blog"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "blog published successfully", Blog: *publishedBlog}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	} else if blog.BlogStatus == storage.BlogStatusPublished && newBlogStatus == storage.BlogStatusArchived {

		archivedBlog, err := h.storage.UpdateBlogStatus(int(blog.Id), newBlogStatus)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool                          `json:"success"`
			Message string                        `json:"message"`
			Blog    storage.BlogWithUserAndTopics `json:"blog"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "blog archived successfully", Blog: *archivedBlog}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	} else if blog.BlogStatus == storage.BlogStatusArchived && newBlogStatus == storage.BlogStatusPublished {

		//	an archived blog means once it was published (so it has topics already)
		publishedBlog, err := h.storage.UpdateBlogStatus(int(blog.Id), newBlogStatus)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool                          `json:"success"`
			Message string                        `json:"message"`
			Blog    storage.BlogWithUserAndTopics `json:"blog"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "blog published successfully", Blog: *publishedBlog}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		writeJSONError(w, fmt.Sprintf("cannot update blog status with current blog status %v", blog.BlogStatus), http.StatusBadRequest)
		return
	}
}

func isArrayContainElement(arr []int, target int) bool {

	for _, val := range arr {
		if val == target {
			return true
		}
	}

	return false
}
