package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"language-learner/db"

	"github.com/gin-gonic/gin"
)

const maxCustomWords = 100

type wordResponse struct {
	ID        int       `json:"id"`
	TopicID   *int      `json:"topicId"`
	Hungarian string    `json:"hungarian"`
	English   string    `json:"english"`
	CreatedAt time.Time `json:"createdAt"`
}

func AddCustomWord(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId"`
		Hungarian string `json:"hungarian"`
		English   string `json:"english"`
		TopicID   *int   `json:"topicId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}
	if req.Hungarian == "" || req.English == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hungarian and english are required"})
		return
	}

	// Check session exists and get current count
	var customWordsCount int
	err := db.DB.QueryRow(
		"SELECT custom_words_count FROM user_sessions WHERE session_id = ?",
		req.SessionID,
	).Scan(&customWordsCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session"})
		return
	}

	if customWordsCount >= maxCustomWords {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":            "Custom words limit reached (100/100)",
			"customWordsCount": customWordsCount,
		})
		return
	}

	// Insert custom word
	now := time.Now()
	result, err := db.DB.Exec(
		"INSERT INTO custom_words (session_id, topic_id, hungarian, english, created_at) VALUES (?, ?, ?, ?, ?)",
		req.SessionID, req.TopicID, req.Hungarian, req.English, now,
	)
	if err != nil {
		slog.Error("Failed to insert custom word", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add custom word"})
		return
	}

	wordID, err := result.LastInsertId()
	if err != nil {
		slog.Error("Failed to get inserted word ID", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add custom word"})
		return
	}

	// Increment custom_words_count
	_, err = db.DB.Exec(
		"UPDATE user_sessions SET custom_words_count = custom_words_count + 1 WHERE session_id = ?",
		req.SessionID,
	)
	if err != nil {
		slog.Error("Failed to update custom_words_count", "error", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        int(wordID) + customWordOffset,
		"hungarian": req.Hungarian,
		"english":   req.English,
		"topicId":   req.TopicID,
		"createdAt": now,
	})
}

func GetCustomWords(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	topicIDStr := c.Query("topicId")

	var rows *sql.Rows
	var err error

	if topicIDStr != "" {
		topicID, parseErr := strconv.Atoi(topicIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topicId"})
			return
		}
		rows, err = db.DB.Query(
			"SELECT id, topic_id, hungarian, english, created_at FROM custom_words WHERE session_id = ? AND topic_id = ? ORDER BY created_at DESC",
			sessionID, topicID,
		)
	} else {
		rows, err = db.DB.Query(
			"SELECT id, topic_id, hungarian, english, created_at FROM custom_words WHERE session_id = ? ORDER BY created_at DESC",
			sessionID,
		)
	}

	if err != nil {
		slog.Error("Failed to fetch custom words", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch custom words"})
		return
	}
	defer rows.Close()

	words := []wordResponse{}

	for rows.Next() {
		var w wordResponse
		var topicID sql.NullInt64
		if err := rows.Scan(&w.ID, &topicID, &w.Hungarian, &w.English, &w.CreatedAt); err != nil {
			slog.Error("Failed to scan custom word", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read custom words"})
			return
		}
		if topicID.Valid {
			tid := int(topicID.Int64)
			w.TopicID = &tid
		}
		w.ID = w.ID + customWordOffset
		words = append(words, w)
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(words),
		"words": words,
	})
}

func DeleteCustomWord(c *gin.Context) {
	wordIDStr := c.Param("wordId")
	sessionID := c.Query("sessionId")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	wordID, err := strconv.Atoi(wordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	// wordId is the offset ID, subtract to get real DB id
	realID := wordID - customWordOffset

	// Verify the session owns this word
	var ownerSessionID string
	err = db.DB.QueryRow(
		"SELECT session_id FROM custom_words WHERE id = ?",
		realID,
	).Scan(&ownerSessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
		return
	}

	if ownerSessionID != sessionID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this word"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM custom_words WHERE id = ?", realID)
	if err != nil {
		slog.Error("Failed to delete custom word", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete word"})
		return
	}

	// Decrement custom_words_count
	_, err = db.DB.Exec(
		"UPDATE user_sessions SET custom_words_count = custom_words_count - 1 WHERE session_id = ?",
		sessionID,
	)
	if err != nil {
		slog.Error("Failed to update custom_words_count", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Word deleted",
		"wordId":  wordID,
	})
}
