package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type WhatsAppMessage struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

func SendWhatsAppMessage(number, message string) error {
	baseURL := os.Getenv("WHATSAPP_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://gate.whapi.cloud" // default, adjust as needed
	}

	url := baseURL + "/messages/text"

	payload := WhatsAppMessage{
		To:   number,
		Body: message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("WHATSAPP_API_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	println(resp.Status + "OTP SENT FROM WHATSAPP")

	return nil
}
