package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func SendTelegramMessage(message string) error {
	payload := TelegramMessagePayload{
		ChatID: TelegramChatId,
		Text:   message,
		ReplyMarkup: InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{},
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf(BotReqURL, TelegramBotToken)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
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
