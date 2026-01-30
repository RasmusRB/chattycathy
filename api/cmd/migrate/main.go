package main

import (
	"os"

	"github.com/chattycathy/api/config"
	"github.com/chattycathy/api/db"
	"github.com/chattycathy/api/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init("info", true)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	// Connect to database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Run migrations
	if err := db.Migrate(database); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run migrations")
	}

	logger.Info().Msg("Migrations completed successfully")
	os.Exit(0)
}
