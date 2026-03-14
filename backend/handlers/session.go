package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"language-learner/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateSession(c *gin.Context) {
	sessionID := uuid.New().String()

	_, err := db.DB.Exec(
		"INSERT INTO user_sessions (session_id, created_at, last_accessed, custom_words_count) VALUES (?, ?, ?, 0)",
		sessionID, time.Now(), time.Now(),
	)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	slog.Info("Session created", "sessionId", sessionID)
	c.JSON(http.StatusCreated, gin.H{"sessionId": sessionID})
}

func ValidateSession(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	var createdAt time.Time
	var customWordsCount int

	err := db.DB.QueryRow(
		"SELECT created_at, custom_words_count FROM user_sessions WHERE session_id = ?",
		sessionID,
	).Scan(&createdAt, &customWordsCount)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":            true,
		"customWordsCount": customWordsCount,
		"createdAt":        createdAt,
	})
}
