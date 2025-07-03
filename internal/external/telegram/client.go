package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
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
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
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

func SendCriticalTelegramMessage(message string) error {
	payload := TelegramMessagePayload{
		ChatID: CriticalTelegramChatId,
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
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
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

func SendVault2Message(message string) error {
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

	url := fmt.Sprintf(BotReqURL, "7734602965:AAERANyQgqr4Lae5u4BFBzdlTDGMc8s9F2s")
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
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
