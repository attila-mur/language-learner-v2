package models

import "time"

type CustomWord struct {
	ID        int       `json:"id" db:"id"`
	SessionID string    `json:"sessionId" db:"session_id"`
	TopicID   *int      `json:"topicId" db:"topic_id"`
	Hungarian string    `json:"hungarian" db:"hungarian"`
	English   string    `json:"english" db:"english"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
