package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

// App represents the application state
type App struct {
	config       Config
	server       *Server
	serverMutex  sync.Mutex
	serverActive bool
}

// NewApp creates a new application instance
func NewApp() *App {
	config, err := LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		config = DefaultConfig()
	}

	return &App{
		config:       config,
		serverActive: false,
	}
}

// Run starts the application
func (a *App) Run() {
	// Start the systray
	systray.Run(a.onReady, a.onExit)
}

// onReady is called when the systray is ready
func (a *App) onReady() {
	// Set icon (using a simple icon for now)
	systray.SetIcon(getIcon())
	systray.SetTitle("OpenRouter Proxy")
	systray.SetTooltip("OpenRouter Proxy for Ollama")

	// Create menu items
	mStatus := systray.AddMenuItem("Status: Stopped", "Server status")
	mStatus.Disable()
	systray.AddSeparator()

	mToggle := systray.AddMenuItem("Start Server", "Start/Stop the proxy server")
	mAPIKey := systray.AddMenuItem("Configure API Key", "Set your OpenRouter API key")
	mModelFilter := systray.AddMenuItem("Edit Model Filter", "Edit the model filter file")
	
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About", "About OpenRouter Proxy")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Start server if enabled in config
	if a.config.ServerEnabled {
		go a.startServer()
	}

	// Handle menu item clicks
	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				a.serverMutex.Lock()
				if a.serverActive {
					a.stopServer()
					mToggle.SetTitle("Start Server")
					mStatus.SetTitle("Status: Stopped")
				} else {
					if !HasAPIKey() {
						a.showAPIKeyDialog()
					}
					
					if HasAPIKey() {
						a.startServer()
						mToggle.SetTitle("Stop Server")
						mStatus.SetTitle("Status: Running")
					}
				}
				a.serverMutex.Unlock()

			case <-mAPIKey.ClickedCh:
				a.showAPIKeyDialog()

			case <-mModelFilter.ClickedCh:
				a.openModelFilter()

			case <-mAbout.ClickedCh:
				a.showAbout()

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

// onExit is called when the systray is exiting
func (a *App) onExit() {
	a.stopServer()
	
	// Save config
	err := SaveConfig(a.config)
	if err != nil {
		slog.Error("Failed to save config", "error", err)
	}
}

// startServer starts the proxy server
func (a *App) startServer() {
	a.serverMutex.Lock()
	defer a.serverMutex.Unlock()

	if a.serverActive {
		return
	}

	apiKey, err := GetAPIKey()
	if err != nil {
		slog.Error("Failed to get API key", "error", err)
		return
	}

	// Create and start the server
	a.server = NewServer(apiKey, a.config.LastUsedModelFilter)
	go a.server.Start()

	a.serverActive = true
	a.config.ServerEnabled = true
	SaveConfig(a.config)

	// Update icon to indicate server is running
	systray.SetIcon(getActiveIcon())
}

// stopServer stops the proxy server
func (a *App) stopServer() {
	a.serverMutex.Lock()
	defer a.serverMutex.Unlock()

	if !a.serverActive || a.server == nil {
		return
	}

	a.server.Stop()
	a.server = nil
	a.serverActive = false
	a.config.ServerEnabled = false
	SaveConfig(a.config)

	// Update icon to indicate server is stopped
	systray.SetIcon(getIcon())
}

// showAPIKeyDialog shows a dialog to configure the API key
func (a *App) showAPIKeyDialog() {
	// For simplicity, we'll use a command-line prompt for now
	// In a real application, you would use a proper GUI dialog
	fmt.Println("Please enter your OpenRouter API key:")
	var apiKey string
	fmt.Scanln(&apiKey)

	if apiKey != "" {
		err := SetAPIKey(apiKey)
		if err != nil {
			slog.Error("Failed to save API key", "error", err)
		} else {
			fmt.Println("API key saved successfully!")
		}
	}
}

// openModelFilter opens the model filter file in the default text editor
func (a *App) openModelFilter() {
	// Ensure the model filter file exists
	if _, err := os.Stat(a.config.LastUsedModelFilter); os.IsNotExist(err) {
		// Create an empty file if it doesn't exist
		file, err := os.Create(a.config.LastUsedModelFilter)
		if err != nil {
			slog.Error("Failed to create model filter file", "error", err)
			return
		}
		file.Close()
	}

	// Open the file in the default text editor
	err := open.Run(a.config.LastUsedModelFilter)
	if err != nil {
		slog.Error("Failed to open model filter file", "error", err)
	}
}

// showAbout shows information about the application
func (a *App) showAbout() {
	message := `OpenRouter Proxy for Ollama

This application provides a proxy server that emulates Ollama's REST API
but forwards requests to OpenRouter.

For more information, visit:
https://github.com/your-username/ollama-openrouter-proxy`

	// For simplicity, just print to console
	// In a real application, you would show a proper dialog
	fmt.Println(message)
}

// getIcon returns the default icon
func getIcon() []byte {
	// This is a placeholder. In a real application, you would include a proper icon.
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0xF3, 0xFF, 0x61, 0x00, 0x00, 0x00,
		0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0B, 0x13, 0x00, 0x00, 0x0B,
		0x13, 0x01, 0x00, 0x9A, 0x9C, 0x18, 0x00, 0x00, 0x00, 0x07, 0x74, 0x49,
		0x4D, 0x45, 0x07, 0xD5, 0x0C, 0x0F, 0x0A, 0x2F, 0x0B, 0x57, 0x77, 0x47,
		0x89, 0x00, 0x00, 0x00, 0x1D, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0x63,
		0x64, 0x60, 0x60, 0xF8, 0xCF, 0x40, 0x21, 0x60, 0x62, 0xA0, 0x10, 0x8C,
		0x1A, 0x30, 0x6A, 0xC0, 0xA8, 0x01, 0x83, 0xD5, 0x00, 0x00, 0x06, 0x10,
		0x00, 0x01, 0x53, 0xFE, 0xD9, 0x3F, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
		0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

// getActiveIcon returns the icon for when the server is active
func getActiveIcon() []byte {
	// This is a placeholder. In a real application, you would include a proper icon.
	return getIcon() // Using the same icon for now
}