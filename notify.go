package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"fmt"

	"database/sql"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type Subscription struct {
	id               int
	createdAt        string
	pushSubscription json.RawMessage
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
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

	message := &Notification{Title: "test title", Body: "hii"}

	err = sendNotifications(message)
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

// https://github.com/nakamauwu/nakama/blob/main/web_push_subscription.go
// Sends notification to all subscriptions in database
func sendNotifications(notification *Notification) error {
	subs, err := getSubscriptions()
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		fmt.Println("there are no active subscriptions")
		return nil
	}

	message, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("could not json marshal notification: %w", err)
	}

	var wg sync.WaitGroup

	for _, sub := range subs {
		wg.Add(1)

		pushSubscription := sub

		go func() {
			defer wg.Done()

			err := sendWebPushNotification(pushSubscription, message)
			if err != nil {
				fmt.Print(err)
			}
		}()
	}

	wg.Wait()

	return nil
}

// Retrieves all subscriptions from database
func getSubscriptions() ([]Subscription, error) {
	query := "SELECT * FROM \"PushSubscriptions\";"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve subscriptions from databse: %w", err)
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription

		err := rows.Scan(&sub.id, &sub.createdAt, &sub.pushSubscription)
		if err != nil {
			continue
		}

		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not sql query iterate over user web push subscriptions: %w", err)
	}

	return subs, nil
}

// Sends push notification to subscription
func sendWebPushNotification(rawSub Subscription, message []byte) error {
	sub := &webpush.Subscription{}
	err := json.Unmarshal([]byte(rawSub.pushSubscription), sub)
	if err != nil {
		return fmt.Errorf("could not json unmarshal web push subscription: %w", err)
	}

	resp, err := webpush.SendNotification(message, sub, &webpush.Options{
		Subscriber:      "mailto:keadr23@gmail.com",
		VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		TTL:             18000, // 5 hours
	})
	if err != nil {
		return fmt.Errorf("could not send web push notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 || resp.StatusCode == 410 {
		fmt.Printf("deleting stale subscription: ID: %v createdAt: %v\n", rawSub.id, rawSub.createdAt)
		err := deleteWebPushSubscription(rawSub.id)
		if err != nil {
			return err
		}
	}

	return nil
}

// Deletes subscription from database
func deleteWebPushSubscription(id int) error {
	query := "DELETE FROM \"PushSubscriptions\" where id = $1;"
	_, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("could not delete subscription from database: %w", err)
	}

	return nil
}
