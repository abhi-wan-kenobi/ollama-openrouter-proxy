package main

import (
	"log/slog"
	"os"
)

func main() {
	// Set up logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Check if API key is provided as command-line argument
	if len(os.Args) > 1 {
		apiKey := os.Args[1]
		err := SetAPIKey(apiKey)
		if err != nil {
			slog.Error("Failed to save API key", "error", err)
			return
		}
		slog.Info("API key saved successfully")
		return
	}

	// Check if API key is provided as environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		err := SetAPIKey(apiKey)
		if err != nil {
			slog.Error("Failed to save API key", "error", err)
		} else {
			slog.Info("API key saved from environment variable")
		}
	}

	// Create and run the application
	app := NewApp()
	app.Run()
}
