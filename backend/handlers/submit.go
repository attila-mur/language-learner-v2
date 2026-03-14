package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"language-learner/db"
	"language-learner/models"

	"github.com/gin-gonic/gin"
)

func SubmitQuiz(c *gin.Context) {
	var req models.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	if len(req.Answers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No answers provided"})
		return
	}

	results := []models.CardResult{}
	score := 0

	for _, answer := range req.Answers {
		var hungarian, english string
		var scanErr error

		if answer.CardID >= customWordOffset {
			// Look up in custom_words table (subtract offset to get real DB id)
			realID := answer.CardID - customWordOffset
			scanErr = db.DB.QueryRow(
				"SELECT hungarian, english FROM custom_words WHERE id = ? AND session_id = ?",
				realID, req.SessionID,
			).Scan(&hungarian, &english)
		} else {
			// Look up in words table
			scanErr = db.DB.QueryRow(
				"SELECT hungarian, english FROM words WHERE id = ?",
				answer.CardID,
			).Scan(&hungarian, &english)
		}

		if scanErr != nil {
			slog.Error("Failed to find card for answer", "cardId", answer.CardID, "error", scanErr)
			continue
		}

		isCorrect := strings.EqualFold(
			strings.TrimSpace(answer.UserAnswer),
			strings.TrimSpace(english),
		)
		if isCorrect {
			score++
		}

		results = append(results, models.CardResult{
			CardID:     answer.CardID,
			Hungarian:  hungarian,
			Correct:    english,
			UserAnswer: answer.UserAnswer,
			IsCorrect:  isCorrect,
		})
	}

	totalCards := len(results)
	percentage := 0
	if totalCards > 0 {
		percentage = (score * 100) / totalCards
	}

	// Store score in database (non-fatal if it fails)
	if _, err := db.DB.Exec(
		"INSERT INTO scores (session_id, topic_id, score, total, date) VALUES (?, ?, ?, ?, ?)",
		req.SessionID, req.TopicID, score, totalCards, time.Now(),
	); err != nil {
		slog.Error("Failed to store score", "error", err)
	}

	// Update last_accessed (non-fatal if it fails)
	if _, err := db.DB.Exec(
		"UPDATE user_sessions SET last_accessed = ? WHERE session_id = ?",
		time.Now(), req.SessionID,
	); err != nil {
		slog.Error("Failed to update last_accessed", "error", err)
	}

	c.JSON(http.StatusOK, models.SubmitResponse{
		Score:      score,
		TotalCards: totalCards,
		Percentage: percentage,
		Results:    results,
	})
}
