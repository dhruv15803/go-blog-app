package storage

import (
	"errors"
	"time"
)

type Topic struct {
	Id        int     `db:"id" json:"id"`
	TopicName string  `db:"topic_name" json:"topic_name"`
	CreatedAt string  `db:"created_at" json:"created_at"`
	UpdatedAt *string `db:"updated_at" json:"updated_at"`
}

func (s *Storage) GetTopicById(topicId int) (*Topic, error) {

	var topic Topic

	query := `SELECT id,topic_name,created_at,updated_at FROM topics WHERE id=$1`

	if err := s.db.QueryRowx(query, topicId).StructScan(&topic); err != nil {
		return nil, err
	}

	return &topic, nil
}

func (s *Storage) GetTopicByTopicName(topicName string) (*Topic, error) {

	var topic Topic

	query := `SELECT id,topic_name,created_at,updated_at FROM topics WHERE topic_name=$1`

	if err := s.db.QueryRowx(query, topicName).StructScan(&topic); err != nil {
		return nil, err
	}

	return &topic, nil
}

func (s *Storage) CreateTopic(topicName string) (*Topic, error) {

	var topic Topic

	query := `INSERT INTO topics(topic_name) VALUES($1) RETURNING id,topic_name,created_at,updated_at`

	if err := s.db.QueryRowx(query, topicName).StructScan(&topic); err != nil {
		return nil, err
	}

	return &topic, nil
}

func (s *Storage) DeleteTopic(topicId int) error {

	query := `DELETE FROM topics WHERE id=$1`

	result, err := s.db.Exec(query, topicId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("topic to delete not found")
	}
	return nil
}

func (s *Storage) UpdateTopicNameById(topicId int, topicName string) (*Topic, error) {

	var topic Topic

	query := `UPDATE topics SET topic_name=$1,updated_at=$2 WHERE id=$3 
RETURNING id,topic_name,created_at,updated_at`

	if err := s.db.QueryRowx(query, topicName, time.Now(), topicId).StructScan(&topic); err != nil {
		return nil, err
	}

	return &topic, nil
}

func (s *Storage) GetTopics(skip int, limit int) ([]Topic, error) {

	var topics []Topic

	query := `SELECT id,topic_name,created_at,updated_at FROM topics 
	ORDER BY created_at DESC 
	LIMIT $1 OFFSET $2`

	rows, err := s.db.Queryx(query, limit, skip)
	if err != nil {
		return []Topic{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic Topic

		if err := rows.StructScan(&topic); err != nil {
			return []Topic{}, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (s *Storage) GetAllTopicsCount() (int, error) {

	var totalCount int

	query := `SELECT COUNT(id) FROM topics`

	if err := s.db.QueryRowx(query).Scan(&totalCount); err != nil {
		return -1, err
	}

	return totalCount, nil
}

func (s *Storage) GetTopicsByTopicNameSearch(topicNameSearchText string, skip int, limit int) ([]Topic, error) {

	var topics []Topic

	query := `SELECT * FROM topics WHERE topic_name LIKE $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3`

	topicNameSearchArg := "%" + topicNameSearchText + "%"

	rows, err := s.db.Queryx(query, topicNameSearchArg, limit, skip)
	if err != nil {
		return []Topic{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic Topic

		if err := rows.StructScan(&topic); err != nil {
			return []Topic{}, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (s *Storage) GetTopicsCountByTopicNameSearch(topicNameSearch string) (int, error) {

	var totalCount int

	query := `SELECT COUNT(id) FROM topics WHERE topic_name LIKE $1`
	topicNameSearchArg := "%" + topicNameSearch + "%"

	if err := s.db.QueryRowx(query, topicNameSearchArg).Scan(&totalCount); err != nil {
		return -1, err
	}

	return totalCount, nil
}
