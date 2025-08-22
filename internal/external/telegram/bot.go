package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Largeb0525/personal-tool/database"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartBot() {
	if TelegramOrderBotToken == "" {
		log.Fatal("telegram bot token is not set")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(TelegramOrderBotToken, opts...)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	log.Println("Telegram bot started")
	b.Start(context.Background())
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := database.GetDB()

	// 加入或移除機器人 加db
	if update.MyChatMember != nil {
		processMyChatMember(db, update.MyChatMember)
		return
	}

	if update.Message != nil {
		// Check if the message is from the allowed chat ID
		if fmt.Sprintf("%d", update.Message.Chat.ID) != OrderTelegramChatId {
			log.Printf("Message from unauthorized chat ID: %d", update.Message.Chat.ID)
			return
		}

		if update.Message.NewChatTitle != "" {
			// 更新聊天室名稱
			processNewChatTitle(db, update.Message.Chat.ID, update.Message.NewChatTitle)
		} else {
			processMessage(ctx, b, update.Message)
		}
	}
}

func processMyChatMember(db *sql.DB, myChatMember *models.ChatMemberUpdated) {
	str, _ := json.Marshal(myChatMember)
	log.Printf("myChatMember: %s", string(str))
	chatID := myChatMember.Chat.ID
	newStatus := myChatMember.NewChatMember.Type

	if newStatus == "member" || newStatus == "administrator" {
		log.Printf("Bot was added to chat %d (%s)", chatID, myChatMember.Chat.Title)
		if err := database.InsertOrUpdateChat(db, chatID, myChatMember.Chat.Title); err != nil {
			log.Printf("Failed to insert or update chat: %v", err)
		}
	} else if newStatus == "left" || newStatus == "kicked" {
		log.Printf("Bot was removed from chat %d", chatID)
		if err := database.DeleteChat(db, chatID); err != nil {
			log.Printf("Failed to delete chat: %v", err)
		}
	}
}

func processNewChatTitle(db *sql.DB, chatID int64, newTitle string) {
	log.Printf("Chat %d title changed to %s", chatID, newTitle)
	if err := database.InsertOrUpdateChat(db, chatID, newTitle); err != nil {
		log.Printf("Failed to update chat title: %v", err)
	}
}

func processMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message.ReplyToMessage != nil && message.Text == "2" {
		// get order by message.ReplyToMessage.Text
		merchantID := "XXXXX"
		if message.ReplyToMessage.Text == "2" {
			merchantID = "OOOOO"
		}
		db := database.GetDB()
		chats, err := database.GetChatByTitle(db, merchantID)
		if err != nil {
			log.Printf("Failed to get all chats: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   "Failed to get chat id in db",
			})
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chats[0].ID,
			// 返回order type or else
			Text: fmt.Sprintf("收到訂單 發送到指定商戶%s", merchantID),
		})
	}
}
