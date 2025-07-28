package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Largeb0525/personal-tool/internal/httpclient"
)

func SendTelegramMessage(message string, chatID string, botToken string) error {
	payload := TelegramMessagePayload{
		ChatID: chatID,
		Text:   message,
		ReplyMarkup: InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{},
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf(BotReqURL, botToken)
	resp, err := httpclient.DefaultClient.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var errResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	desc := "No error description available"
	if d, ok := errResp["description"].(string); ok {
		desc = d
	}
	log.Printf("Failed to send message: %d, %s\n", resp.StatusCode, desc)
	return fmt.Errorf("telegram error: %s", desc)
}
