package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/spf13/viper"
)

func StartBot() {
	botToken := viper.GetString("andy.telegram.bot_token")
	if botToken == "" {
		log.Fatal("telegram bot token is not set")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(botToken, opts...)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	log.Println("Telegram bot started")
	b.Start(context.Background())
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := viper.GetString("andy.telegram.chat_id")
	if chatID == "" {
		log.Println("telegram chat id is not set")
		return
	}

	if fmt.Sprintf("%d", update.Message.Chat.ID) != chatID {
		return
	}

	if update.Message.ReplyToMessage != nil && update.Message.Text == "test" {
		replyText := fmt.Sprintf("[%d, \"%s\"]", update.Message.Chat.ID, update.Message.Chat.Title)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   replyText,
		})
	}
}
