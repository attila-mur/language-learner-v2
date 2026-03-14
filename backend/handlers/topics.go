package handlers

import (
	"log/slog"
	"net/http"

	"language-learner/db"

	"github.com/gin-gonic/gin"
)

const maxCustomTopics = 4

func CreateCustomTopic(c *gin.Context) {
	var req struct {
		SessionID   string `json:"sessionId"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Verify session exists
	var exists int
	err := db.DB.QueryRow(
		"SELECT COUNT(*) FROM user_sessions WHERE session_id = ?", req.SessionID,
	).Scan(&exists)
	if err != nil || exists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session"})
		return
	}

	// Enforce custom topic limit
	var topicCount int
	if err := db.DB.QueryRow(
		"SELECT COUNT(*) FROM topics WHERE session_id = ?", req.SessionID,
	).Scan(&topicCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check topic count"})
		return
	}
	if topicCount >= maxCustomTopics {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Custom topic limit reached (4/4)",
			"topicCount":  topicCount,
			"topicLimit":  maxCustomTopics,
		})
		return
	}

	result, err := db.DB.Exec(
		"INSERT INTO topics (name, session_id, description) VALUES (?, ?, ?)",
		req.Name, req.SessionID, req.Description,
	)
	if err != nil {
		slog.Error("Failed to create custom topic", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create topic"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":          id,
		"name":        req.Name,
		"description": req.Description,
		"isCustom":    true,
	})
}
