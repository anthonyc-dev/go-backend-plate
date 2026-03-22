package services

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ExpoMessage struct {
	To    string      `json:"to"`
	Title string      `json:"title"`
	Body  string      `json:"body"`
	Data  interface{} `json:"data,omitempty"`
}

// SendPush handles batching (max 100 per request)
func SendPush(messages []ExpoMessage) error {
	if len(messages) == 0 {
		return nil
	}

	chunkSize := 100

	for i := 0; i < len(messages); i += chunkSize {
		end := i + chunkSize
		if end > len(messages) {
			end = len(messages)
		}

		chunk := messages[i:end]

		jsonData, _ := json.Marshal(chunk)

		req, _ := http.NewRequest(
			"POST",
			"https://exp.host/--/api/v2/push/send",
			bytes.NewBuffer(jsonData),
		)

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		_, err := client.Do(req)
		if err != nil {
			return err
		}
	}

	return nil
}
