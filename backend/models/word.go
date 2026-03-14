package models

import "time"

type Topic struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsCustom    bool   `json:"isCustom"`
}

type Word struct {
	ID        int       `json:"id" db:"id"`
	TopicID   int       `json:"topicId" db:"topic_id"`
	Hungarian string    `json:"hungarian" db:"hungarian"`
	English   string    `json:"english" db:"english"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
