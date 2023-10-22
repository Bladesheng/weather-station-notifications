package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Fetches notification content from server
func GetNotification() (*Notification, error) {
	url := "https://weather-station-backend.fly.dev"
	// url := "http://localhost:8080"

	resp, err := http.Get(url + "/api/forecast/notification")
	if err != nil {
		return nil, fmt.Errorf("could not get forecast: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read forecast body: %w", err)
	}

	notification := &Notification{}
	err = json.Unmarshal(body, notification)
	if err != nil {
		return nil, fmt.Errorf("could not json unmarshal notification content: %w", err)
	}

	return notification, nil
}
