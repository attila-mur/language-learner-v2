package models

import "time"

type UserSession struct {
	SessionID        string    `json:"sessionId" db:"session_id"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	LastAccessed     time.Time `json:"lastAccessed" db:"last_accessed"`
	CustomWordsCount int       `json:"customWordsCount" db:"custom_words_count"`
}
