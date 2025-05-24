package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shawty/internal/config"
	"shawty/internal/handler"
	"shawty/internal/service"
	"shawty/internal/store"
	"syscall"
	"time"
)

func main() {
	// Load application configuration
	dbCfg := config.LoadConfig()

	// Connect to MongoDB
	dbClient, err := config.ConnectDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err = dbClient.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect MongoDB client: %v", err)
		} else {
			fmt.Println("Disconnected from MongoDB.")
		}
	}()

	// Initialize store
	urlStore := store.NewMongoUrlStore(dbClient, dbCfg.DBName, dbCfg.CollectionName)

	// This is a good practice to do on startup.
	ctx, cancelIdx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelIdx()
	if err := urlStore.EnsureIndexes(ctx); err != nil {
		log.Fatalf("Failed to ensure database indexes: %v", err)
	}

	// Initialize service
	urlSvc := service.NewUrlService(urlStore)

	// Initialize HTTP handler
	urlHandler := handler.NewURLHandler(urlSvc)

	// Setup HTTP server and routes
	mux := http.NewServeMux()
	urlHandler.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
		log.Printf("PORT environment variable not set, using default %s", port)
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(mux),
		// Good practice: add timeouts to avoid resource exhaustion.
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// corsMiddleware adds necessary CORS headers to each request.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set allowed origin. For production, you might want to make this configurable
		// and more restrictive than "*". For development, "http://localhost:3000" is specific.
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		// Set allowed methods
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		// Set allowed headers
		// "Content-Type" is important for your frontend's POST request.
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests (OPTIONS method)
		// Browsers send an OPTIONS request first to check if the actual request is safe to send.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass control to the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
