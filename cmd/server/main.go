package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/khahanv2/account-login-checker/internal/handlers"
	"github.com/khahanv2/account-login-checker/internal/logger"
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Create required directories
	createRequiredDirectories()

	// Start WebSocket server (same port as HTTP server)
	logger.StartWebSocketServer("8080")

	// Setup HTTP router
	router := handlers.SetupRouter()

	// Log application start
	fmt.Println("Starting Account Login Checker Server...")
	fmt.Println("Web Interface: http://localhost:8080")
	fmt.Println("WebSocket: ws://localhost:8080/ws")
	fmt.Println("Press Ctrl+C to exit")

	// Start HTTP server
	err := router.Run(":8080")
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}

// createRequiredDirectories creates the necessary directories for the application
func createRequiredDirectories() {
	dirs := []string{"temp", "results", "static/js", "static/css", "static/images"}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory '%s': %s\n", dir, err)
		}
	}
}