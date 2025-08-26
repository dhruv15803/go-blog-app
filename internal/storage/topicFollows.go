package storage

import "errors"

type TopicFollow struct {
	UserId     int    `db:"user_id" json:"user_id"`
	TopicId    int    `db:"topic_id" json:"topic_id"`
	FollowedAt string `db:"followed_at" json:"followed_at"`
}

func (s *Storage) GetTopicFollow(userId int, topicId int) (*TopicFollow, error) {

	var topicFollow TopicFollow

	query := `SELECT user_id,user_id, topic_id, followed_at 
	FROM topic_follows WHERE user_id=$1 AND topic_id=$2`

	if err := s.db.QueryRowx(query, userId, topicId).StructScan(&topicFollow); err != nil {
		return nil, err
	}

	return &topicFollow, nil
}

func (s *Storage) CreateTopicFollow(userId int, topicId int) (*TopicFollow, error) {

	var topicFollow TopicFollow

	query := `INSERT INTO topic_follows(user_id,topic_id) VALUES($1,$2) RETURNING 	
	user_id,topic_id,followed_at`

	if err := s.db.QueryRowx(query, userId, topicId).StructScan(&topicFollow); err != nil {
		return nil, err
	}

	return &topicFollow, nil
}

func (s *Storage) RemoveTopicFollow(userId int, topicId int) error {

	query := `DELETE FROM topic_follows WHERE user_id=$1 AND topic_id=$2`

	result, err := s.db.Exec(query, userId, topicId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("failed to remove topic follow")
	}

	return nil
}
