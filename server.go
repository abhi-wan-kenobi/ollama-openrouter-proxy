package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

// Server encapsulates the proxy server functionality
type Server struct {
	apiKey      string
	modelFilter string
	router      *gin.Engine
	httpServer  *http.Server
	provider    *OpenrouterProvider
	filterMap   map[string]struct{}
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewServer creates a new server instance
func NewServer(apiKey, modelFilter string) *Server {
	return &Server{
		apiKey:      apiKey,
		modelFilter: modelFilter,
		stopCh:      make(chan struct{}),
	}
}

// Start starts the proxy server
func (s *Server) Start() {
	s.wg.Add(1)
	defer s.wg.Done()

	// Initialize the provider
	s.provider = NewOpenrouterProvider(s.apiKey)

	// Load model filter
	filter, err := s.loadModelFilter(s.modelFilter)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("models-filter file not found. Skipping model filtering.")
			s.filterMap = make(map[string]struct{})
		} else {
			slog.Error("Error loading models filter", "Error", err)
			return
		}
	} else {
		s.filterMap = filter
		slog.Info("Loaded models from filter:")
		for model := range s.filterMap {
			slog.Info(" - " + model)
		}
	}

	// Set up the router
	s.router = gin.Default()
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    ":11434",
		Handler: s.router,
	}

	// Start the server
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	slog.Info("Server started on port 11434")

	// Wait for stop signal
	<-s.stopCh
}

// Stop stops the proxy server
func (s *Server) Stop() {
	if s.httpServer != nil {
		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Shutdown the server
		if err := s.httpServer.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown error", "error", err)
		}

		// Signal the Start method to return
		close(s.stopCh)

		// Wait for the Start method to complete
		s.wg.Wait()

		slog.Info("Server stopped")
	}
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Ollama is running")
	})
	s.router.HEAD("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	s.router.GET("/api/tags", func(c *gin.Context) {
		models, err := s.provider.GetModels()
		if err != nil {
			slog.Error("Error getting models", "Error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		filter := s.filterMap
		// Construct a new array of model objects with extra fields
		newModels := make([]map[string]interface{}, 0, len(models))
		for _, m := range models {
			// If filter is not empty, check if model is in filter
			if len(filter) > 0 {
				if _, ok := filter[m.Model]; !ok {
					continue
				}
			}
			newModels = append(newModels, map[string]interface{}{
				"name":        m.Name,
				"model":       m.Model,
				"modified_at": m.ModifiedAt,
				"size":        270898672,
				"digest":      "9077fe9d2ae1a4a41a868836b56b8163731a8fe16621397028c2c76f838c6907",
				"details":     m.Details,
			})
		}

		c.JSON(http.StatusOK, gin.H{"models": newModels})
	})

	s.router.POST("/api/show", func(c *gin.Context) {
		var request map[string]string
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		modelName := request["name"]
		if modelName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
			return
		}

		details, err := s.provider.GetModelDetails(modelName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, details)
	})

	s.router.POST("/api/chat", func(c *gin.Context) {
		var request struct {
			Model    string                         `json:"model"`
			Messages []openai.ChatCompletionMessage `json:"messages"`
			Stream   *bool                          `json:"stream"`
		}

		// Parse the JSON request
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		// Determine if streaming is requested (default true for /api/chat)
		streamRequested := true
		if request.Stream != nil {
			streamRequested = *request.Stream
		}

		// Handle non-streaming response
		if !streamRequested {
			fullModelName, err := s.provider.GetFullModelName(request.Model)
			if err != nil {
				slog.Error("Error getting full model name", "Error", err)
				// Ollama returns 404 for invalid model names
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			// Call Chat to get the complete response
			response, err := s.provider.Chat(request.Messages, fullModelName)
			if err != nil {
				slog.Error("Failed to get chat response", "Error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Format the response according to Ollama's format
			if len(response.Choices) == 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "No response from model"})
				return
			}

			// Extract the content from the response
			content := ""
			if len(response.Choices) > 0 && response.Choices[0].Message.Content != "" {
				content = response.Choices[0].Message.Content
			}

			// Get finish reason, default to "stop" if not provided
			finishReason := "stop"
			if response.Choices[0].FinishReason != "" {
				finishReason = string(response.Choices[0].FinishReason)
			}

			// Create Ollama-compatible response
			ollamaResponse := map[string]interface{}{
				"model":             fullModelName,
				"created_at":        time.Now().Format(time.RFC3339),
				"message": map[string]string{
					"role":    "assistant",
					"content": content,
				},
				"done":              true,
				"finish_reason":     finishReason,
				"total_duration":    response.Usage.TotalTokens * 10, // Approximate duration based on token count
				"load_duration":     0,
				"prompt_eval_count": response.Usage.PromptTokens,
				"eval_count":        response.Usage.CompletionTokens,
				"eval_duration":     response.Usage.CompletionTokens * 10, // Approximate duration based on token count
			}

			c.JSON(http.StatusOK, ollamaResponse)
			return
		}

		slog.Info("Requested model", "model", request.Model)
		fullModelName, err := s.provider.GetFullModelName(request.Model)
		if err != nil {
			slog.Error("Error getting full model name", "Error", err, "model", request.Model)
			// Ollama returns 404 for invalid model names
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		slog.Info("Using model", "fullModelName", fullModelName)

		// Call ChatStream to get the stream
		stream, err := s.provider.ChatStream(request.Messages, fullModelName)
		if err != nil {
			slog.Error("Failed to create stream", "Error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer stream.Close() // Ensure stream closure

		// Set headers for Newline Delimited JSON
		c.Writer.Header().Set("Content-Type", "application/x-ndjson")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		w := c.Writer
		flusher, ok := w.(http.Flusher)
		if !ok {
			slog.Error("Expected http.ResponseWriter to be an http.Flusher")
			return
		}

		var lastFinishReason string

		// Stream responses back to the client
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// End of stream from the backend provider
				break
			}
			if err != nil {
				slog.Error("Backend stream error", "Error", err)
				// Try to send error in NDJSON format
				errorMsg := map[string]string{"error": "Stream error: " + err.Error()}
				errorJson, _ := json.Marshal(errorMsg)
				fmt.Fprintf(w, "%s\n", string(errorJson))
				flusher.Flush()
				return
			}

			// Save finish reason if present in chunk
			if len(response.Choices) > 0 && response.Choices[0].FinishReason != "" {
				lastFinishReason = string(response.Choices[0].FinishReason)
			}

			// Build JSON response structure for intermediate chunks
			responseJSON := map[string]interface{}{
				"model":      fullModelName,
				"created_at": time.Now().Format(time.RFC3339),
				"message": map[string]string{
					"role":    "assistant",
					"content": response.Choices[0].Delta.Content,
				},
				"done": false,
			}

			// Marshal JSON
			jsonData, err := json.Marshal(responseJSON)
			if err != nil {
				slog.Error("Error marshaling intermediate response JSON", "Error", err)
				return
			}

			// Send JSON object followed by a newline
			fmt.Fprintf(w, "%s\n", string(jsonData))

			// Flush data to send it immediately
			flusher.Flush()
		}

		// Set finish reason (default to 'stop')
		if lastFinishReason == "" {
			lastFinishReason = "stop"
		}

		// Send final message with done=true
		finalResponse := map[string]interface{}{
			"model":             fullModelName,
			"created_at":        time.Now().Format(time.RFC3339),
			"message": map[string]string{
				"role":    "assistant",
				"content": "",
			},
			"done":              true,
			"finish_reason":     lastFinishReason,
			"total_duration":    0,
			"load_duration":     0,
			"prompt_eval_count": 0,
			"eval_count":        0,
			"eval_duration":     0,
		}

		finalJsonData, err := json.Marshal(finalResponse)
		if err != nil {
			slog.Error("Error marshaling final response JSON", "Error", err)
			return
		}

		// Send final JSON object + newline
		fmt.Fprintf(w, "%s\n", string(finalJsonData))
		flusher.Flush()
	})
}

// loadModelFilter loads the model filter from a file
func (s *Server) loadModelFilter(path string) (map[string]struct{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	filter := make(map[string]struct{})

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			filter[line] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return filter, nil
}