package handlers

import (
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"language-learner/db"
	"language-learner/models"

	"github.com/gin-gonic/gin"
)

const customWordOffset = 1_000_000
const maxQuizCards = 10

func GetTopics(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, name, description FROM topics ORDER BY id")
	if err != nil {
		slog.Error("Failed to fetch topics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}
	defer rows.Close()

	topics := []models.Topic{}
	for rows.Next() {
		var t models.Topic
		if err := rows.Scan(&t.ID, &t.Name, &t.Description); err != nil {
			slog.Error("Failed to scan topic", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read topics"})
			return
		}
		topics = append(topics, t)
	}

	c.JSON(http.StatusOK, topics)
}

func GetQuiz(c *gin.Context) {
	topicIDStr := c.Param("topicId")
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic ID"})
		return
	}

	sessionID := c.Query("sessionId")
	includeCustom := c.DefaultQuery("includeCustom", "true") != "false"

	// Get topic info
	var topicName string
	err = db.DB.QueryRow("SELECT name FROM topics WHERE id = ?", topicID).Scan(&topicName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	// Get seeded words for this topic
	rows, err := db.DB.Query(
		"SELECT id, hungarian FROM words WHERE topic_id = ?",
		topicID,
	)
	if err != nil {
		slog.Error("Failed to fetch words", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch words"})
		return
	}
	defer rows.Close()

	cards := []models.QuizCard{}
	for rows.Next() {
		var card models.QuizCard
		if err := rows.Scan(&card.ID, &card.Hungarian); err != nil {
			slog.Error("Failed to scan word", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read words"})
			return
		}
		card.English = nil
		card.IsCustom = false
		cards = append(cards, card)
	}

	// Include custom words if requested and session is provided
	if includeCustom && sessionID != "" {
		customRows, err := db.DB.Query(
			"SELECT id, hungarian FROM custom_words WHERE session_id = ? AND (topic_id = ? OR topic_id IS NULL)",
			sessionID, topicID,
		)
		if err != nil {
			slog.Error("Failed to fetch custom words", "error", err)
			// Don't fail, just skip custom words
		} else {
			defer customRows.Close()
			for customRows.Next() {
				var card models.QuizCard
				var dbID int
				if err := customRows.Scan(&dbID, &card.Hungarian); err != nil {
					slog.Error("Failed to scan custom word", "error", err)
					continue
				}
				card.ID = dbID + customWordOffset
				card.English = nil
				card.IsCustom = true
				cards = append(cards, card)
			}
		}
	}

	// Shuffle and limit to maxQuizCards
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})

	if len(cards) > maxQuizCards {
		cards = cards[:maxQuizCards]
	}

	c.JSON(http.StatusOK, models.QuizResponse{
		TopicID:    topicID,
		TopicName:  topicName,
		TotalCards: len(cards),
		Cards:      cards,
	})
}
