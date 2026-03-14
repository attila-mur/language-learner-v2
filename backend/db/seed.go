package db

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
)

type seedTopic struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Words       []seedWord `json:"words"`
}

type seedWord struct {
	Hungarian string `json:"hungarian"`
	English   string `json:"english"`
}

func SeedDatabase(dataPath string) error {
	// Check if already seeded
	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM topics").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		slog.Info("Database already seeded, skipping", "topics", count)
		return nil
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}

	var topics []seedTopic
	if err := json.Unmarshal(data, &topics); err != nil {
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	if insertErr := insertTopics(tx, topics); insertErr != nil {
		tx.Rollback()
		return insertErr
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	slog.Info("Database seeding complete", "topics", len(topics))
	return nil
}

func insertTopics(tx *sql.Tx, topics []seedTopic) error {
	for _, topic := range topics {
		result, err := tx.Exec(
			"INSERT INTO topics (name, description) VALUES (?, ?)",
			topic.Name, topic.Description,
		)
		if err != nil {
			return err
		}

		topicID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		for _, word := range topic.Words {
			if _, err := tx.Exec(
				"INSERT INTO words (topic_id, hungarian, english) VALUES (?, ?, ?)",
				topicID, word.Hungarian, word.English,
			); err != nil {
				return err
			}
		}

		slog.Info("Seeded topic", "name", topic.Name, "words", len(topic.Words))
	}
	return nil
}
