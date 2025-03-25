package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hari4698/hardinfinity/internal/api"
	"github.com/hari4698/hardinfinity/internal/db"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initilize database: %v", err)
	}
	defer db.Close()
	
	server := api.NewServer()
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	fmt.Println("Server started successfully")
	
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	fmt.Println("Shutting down server...")
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	fmt.Println("Server stopped succesfully")
}