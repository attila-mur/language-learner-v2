package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"language-learner/db"
	"language-learner/handlers"
	"language-learner/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Configuration from environment
	dbPath := getEnv("DATABASE_PATH", "./quiz.db")
	frontendPath := getEnv("FRONTEND_PATH", "../frontend")
	port := getEnv("PORT", "8080")

	slog.Info("Starting language learner", "port", port, "dbPath", dbPath, "frontendPath", frontendPath)

	// Initialize database
	if err := db.InitDB(dbPath); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.DB.Close()

	// Seed database
	dataPath := getEnv("DATA_PATH", "./data/words.json")
	if err := db.SeedDatabase(dataPath); err != nil {
		slog.Error("Failed to seed database", "error", err)
		// Non-fatal: continue without seeding
	}

	// Set up router
	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORS())

	// Serve frontend static files BEFORE API routes
	router.StaticFile("/", frontendPath+"/index.html")
	router.StaticFile("/style.css", frontendPath+"/style.css")
	router.StaticFile("/app.js", frontendPath+"/app.js")
	router.StaticFile("/pages.js", frontendPath+"/pages.js")
	router.StaticFile("/utils.js", frontendPath+"/utils.js")

	// API routes
	api := router.Group("/api")
	{
		// Session
		api.POST("/session", handlers.CreateSession)
		api.GET("/session/validate", handlers.ValidateSession)

		// Topics & Quiz
		api.GET("/topics", handlers.GetTopics)
		api.POST("/topics", handlers.CreateCustomTopic)
		api.GET("/quiz/:topicId", handlers.GetQuiz)

		// Submit
		api.POST("/submit", handlers.SubmitQuiz)

		// Custom words
		api.POST("/custom-words", handlers.AddCustomWord)
		api.GET("/custom-words", handlers.GetCustomWords)
		api.DELETE("/custom-words/:wordId", handlers.DeleteCustomWord)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		slog.Info("Server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
