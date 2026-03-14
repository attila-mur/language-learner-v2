package handlers

import (
	"database/sql"
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
	sessionID := c.Query("sessionId")

	var (
		rows *sql.Rows
		err  error
	)

	if sessionID != "" {
		// Return seeded topics + this session's custom topics
		rows, err = db.DB.Query(
			`SELECT id, name, COALESCE(description,''), session_id IS NOT NULL
			 FROM topics
			 WHERE session_id IS NULL OR session_id = ?
			 ORDER BY session_id IS NOT NULL, id`,
			sessionID,
		)
	} else {
		rows, err = db.DB.Query(
			`SELECT id, name, COALESCE(description,''), 0
			 FROM topics
			 WHERE session_id IS NULL
			 ORDER BY id`,
		)
	}

	if err != nil {
		slog.Error("Failed to fetch topics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}
	defer rows.Close()

	topics := []models.Topic{}
	for rows.Next() {
		var t models.Topic
		var isCustomInt int
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &isCustomInt); err != nil {
			slog.Error("Failed to scan topic", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read topics"})
			return
		}
		t.IsCustom = isCustomInt != 0
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

	// Get topic info — check whether it's a custom topic
	var topicName string
	var topicSessionID sql.NullString
	err = db.DB.QueryRow(
		"SELECT name, session_id FROM topics WHERE id = ?", topicID,
	).Scan(&topicName, &topicSessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	isCustomTopic := topicSessionID.Valid
	cards := []models.QuizCard{}

	if isCustomTopic {
		// Custom topic: only pull from custom_words for this session
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required for custom topics"})
			return
		}
		rows, err := db.DB.Query(
			"SELECT id, hungarian FROM custom_words WHERE topic_id = ? AND session_id = ?",
			topicID, sessionID,
		)
		if err != nil {
			slog.Error("Failed to fetch custom topic words", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch words"})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var dbID int
			var card models.QuizCard
			if err := rows.Scan(&dbID, &card.Hungarian); err != nil {
				slog.Error("Failed to scan word", "error", err)
				continue
			}
			card.ID = dbID + customWordOffset
			card.English = nil
			card.IsCustom = true
			cards = append(cards, card)
		}
	} else {
		// Seeded topic: pull from words table
		rows, err := db.DB.Query(
			"SELECT id, hungarian FROM words WHERE topic_id = ?", topicID,
		)
		if err != nil {
			slog.Error("Failed to fetch words", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch words"})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var card models.QuizCard
			if err := rows.Scan(&card.ID, &card.Hungarian); err != nil {
				slog.Error("Failed to scan word", "error", err)
				continue
			}
			card.English = nil
			card.IsCustom = false
			cards = append(cards, card)
		}

		// Also include this session's custom words for the seeded topic
		if includeCustom && sessionID != "" {
			customRows, err := db.DB.Query(
				"SELECT id, hungarian FROM custom_words WHERE session_id = ? AND topic_id = ?",
				sessionID, topicID,
			)
			if err != nil {
				slog.Error("Failed to fetch custom words for topic", "error", err)
			} else {
				defer customRows.Close()
				for customRows.Next() {
					var dbID int
					var card models.QuizCard
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
