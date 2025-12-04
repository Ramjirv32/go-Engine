package main

import (
	"log"
	"net/http"
	"os"

	"gobackend/config"
	"gobackend/routes"
)

func main() {
	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	log.Printf("🌍 Environment: %s\n", env)
	log.Printf("📍 Port: %s\n", port)

	// Connect to database
	if err := config.ConnectDatabase(); err != nil {
		log.Fatal("❌ MongoDB connection failed:", err)
	}
	defer config.DisconnectDatabase()

	// Setup routes
	router := routes.SetupRoutes()

	// Log startup info
	log.Printf("🚀 Go Server running on http://localhost:%s\n", port)
	log.Printf("📊 API Health Check: http://localhost:%s/api/health\n", port)

	// Start server
	if env == "production" {
		log.Printf("✅ Server started in production mode")
	} else {
		log.Printf("🔧 Server started in development mode")
	}

	log.Fatal(http.ListenAndServe(":"+port, router))
}
