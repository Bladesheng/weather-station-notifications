package main

import (
	"encoding/json"
	"log"
	"os"

	"fmt"

	"database/sql"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type Subscription struct {
	id               int
	createdAt        string
	pushSubscription json.RawMessage
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon"`
}

var DB *sql.DB

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	err = connectDB()
	if err != nil {
		log.Fatal(err)
	}

	notification, err := GetNotification()
	if err != nil {
		log.Fatal(err)
	}

	err = SendNotifications(notification)
	if err != nil {
		log.Fatal(err)
	}

}

// Loads environment variables from .env file if not in production
func loadEnv() error {
	env := os.Getenv("NODE_ENV")
	if env == "production" {
		return nil
	}

	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("could not load .env file: %w", err)
	}

	return nil
}

func connectDB() error {
	connStr := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	DB = db

	return nil
}
