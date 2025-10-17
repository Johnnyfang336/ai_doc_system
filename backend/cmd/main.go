package main

import (
	"log"
	
	"ai-doc-system/internal/api"
	"ai-doc-system/internal/config"
	"ai-doc-system/internal/database"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	
	// Run database migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	
	// Setup routes
	router := api.SetupRouter(db, cfg.JWTSecret)
	
	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}