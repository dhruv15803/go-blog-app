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

type CreateTopicRequest struct {
	TopicName string `json:"topic_name"`
}

type UpdateTopicRequest struct {
	TopicName string `json:"topic_name"`
}

// admin route
func (h *Handler) CreateTopicHandler(w http.ResponseWriter, r *http.Request) {

	var createTopicPayload CreateTopicRequest

	if err := json.NewDecoder(r.Body).Decode(&createTopicPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	topicName := strings.ToLower(strings.TrimSpace(createTopicPayload.TopicName))

	if topicName == "" {
		writeJSONError(w, "topic name is required", http.StatusBadRequest)
		return
	}

	//	check if topic name already exists
	existingTopic, err := h.storage.GetTopicByTopicName(topicName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if existingTopic != nil {
		writeJSONError(w, "topic already exists", http.StatusBadRequest)
		return
	}

	newTopic, err := h.storage.CreateTopic(topicName)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool          `json:"success"`
		Topic   storage.Topic `json:"topic"`
	}

	if err := writeJSON(w, Response{Success: true, Topic: *newTopic}, http.StatusCreated); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateTopicHandler(w http.ResponseWriter, r *http.Request) {

	var updateTopicPayload UpdateTopicRequest

	topicId, err := strconv.ParseInt(chi.URLParam(r, "topicId"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid request param topicId", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&updateTopicPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newTopicName := strings.ToLower(strings.TrimSpace(updateTopicPayload.TopicName))

	if newTopicName == "" {
		writeJSONError(w, "topic name is required", http.StatusBadRequest)
		return
	}

	topic, err := h.storage.GetTopicById(int(topicId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "topic not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	//	the new topic name should not already exist in the topics , if it does then the topic to be updated
	// should have the new topic name

	topicWithNewTopicName, err := h.storage.GetTopicByTopicName(newTopicName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if topicWithNewTopicName == nil || (topicWithNewTopicName != nil && topicWithNewTopicName.Id == topic.Id) {

		updatedTopic, err := h.storage.UpdateTopicNameById(int(topicId), newTopicName)
		if err != nil {
			log.Printf("failed to update topic name: %v\n", err)
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool          `json:"success"`
			Topic   storage.Topic `json:"topic"`
		}

		if err := writeJSON(w, Response{Success: true, Topic: *updatedTopic}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	} else {
		writeJSONError(w, "topic with topic name already exists", http.StatusBadRequest)
		return
	}
}

// admin route
// {topicId} -> request param
func (h *Handler) DeleteTopicHandler(w http.ResponseWriter, r *http.Request) {

	topicId, err := strconv.ParseInt(chi.URLParam(r, "topicId"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid request param topicId", http.StatusBadRequest)
		return
	}

	topic, err := h.storage.GetTopicById(int(topicId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "topic not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	//	delete this topic

	if err = h.storage.DeleteTopic(topic.Id); err != nil {
		log.Printf("failed to delete topic :%v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "topic deleted"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetTopicsHandler(w http.ResponseWriter, r *http.Request) {

	searchByTopicName := true
	searchText := ""

	if r.URL.Query().Get("search") == "" {
		searchByTopicName = false
	} else {
		searchText = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	}

	limitNum, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid query param limit", http.StatusBadRequest)
		return
	}
	pageNum, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		writeJSONError(w, "invalid query param page", http.StatusBadRequest)
		return
	}

	skip := int(pageNum*limitNum - limitNum)

	//	offset = skip
	// limit = limit

	if searchByTopicName {

		//	query fill first filter down to topic names that contain the search text
		//	then pagination
		topics, err := h.storage.GetTopicsByTopicNameSearch(searchText, skip, int(limitNum))
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		totalTopicsCount, err := h.storage.GetTopicsCountByTopicNameSearch(searchText)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		noOfPages := int(math.Ceil(float64(totalTopicsCount) / float64(limitNum)))

		type Response struct {
			Success   bool            `json:"success"`
			Topics    []storage.Topic `json:"topics"`
			NoOfPages int             `json:"no_of_pages"`
		}

		if err := writeJSON(w, Response{Success: true, Topics: topics, NoOfPages: noOfPages}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	} else {
		topics, err := h.storage.GetTopics(skip, int(limitNum))
		if err != nil {
			log.Printf("failed to get topics: %v\n", err)
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		totalTopicsCount, err := h.storage.GetAllTopicsCount()
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		noOfPages := int(math.Ceil(float64(totalTopicsCount / int(limitNum))))

		type Response struct {
			Success   bool            `json:"success"`
			Topics    []storage.Topic `json:"topics"`
			NoOfPages int             `json:"no_of_pages"`
		}

		if err := writeJSON(w, Response{Success: true, Topics: topics, NoOfPages: noOfPages}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}
