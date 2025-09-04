package andy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Largeb0525/personal-tool/database"
	"github.com/Largeb0525/personal-tool/internal/external/telegram"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartBot() {
	if telegram.TelegramOrderBotToken == "" {
		log.Fatal("telegram bot token is not set")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(telegram.TelegramOrderBotToken, opts...)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	log.Println("Telegram bot started")
	b.Start(context.Background())
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// 加入或移除機器人 加db
	if update.MyChatMember != nil {
		processMyChatMember(update.MyChatMember)
		return
	}

	if update.Message != nil {
		if update.Message.NewChatTitle != "" {
			// 更新聊天室名稱
			processNewChatTitle(update.Message.Chat.ID, update.Message.NewChatTitle)
		} else {
			processMessage(ctx, b, update.Message)
		}
	}
}

func processMyChatMember(myChatMember *models.ChatMemberUpdated) {
	db := database.GetDB()
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

func processNewChatTitle(chatID int64, newTitle string) {
	db := database.GetDB()
	log.Printf("Chat %d title changed to %s", chatID, newTitle)
	if err := database.InsertOrUpdateChat(db, chatID, newTitle); err != nil {
		log.Printf("Failed to update chat title: %v", err)
	}
}

func processMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	textArr := strings.Split(message.Text, " ")
	if message.ReplyToMessage != nil {
		queryType := ""
		orderId := ""
		if textArr[0] == "2" {
			queryType = "merchant_order_id"
			orderId = message.ReplyToMessage.Text
			// if the reply message contains a photo or file
			if message.ReplyToMessage.Caption != "" {
				orderId = message.ReplyToMessage.Caption
			}
			if len(textArr) > 1 {
				orderId = textArr[1]
			}
		} else if textArr[0] == "3" {
			queryType = "search"
			if len(textArr) > 1 {
				orderId = textArr[1]
			} else {
				return
			}
		} else {
			return
		}

		orderInfo, err := getIndiaOrder(orderId, queryType)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   "err: " + err.Error(),
			})
			return
		}

		msg := ""
		var chatId int64
		db := database.GetDB()
		switch orderInfo.OrderStatus {
		case "已完成":
			chats, err := database.GetChatByTitle(db, orderInfo.CustomerUsername)
			if err != nil {
				log.Printf("Failed to get err: %v", err)
			}
			if len(chats) == 0 {
				chatId = message.Chat.ID
				msg = fmt.Sprintf("Cannot find the target chat room\nPlease remind customer %s\n", orderInfo.AdvertiserUsername)
			} else {
				chatId = chats[0].ID
			}
			msg += fmt.Sprintf("Order %s completed.", orderInfo.MerchantOrderId)

		case "已取消":
			chats, err := database.GetChatByTitle(db, orderInfo.AdvertiserUsername)
			if err != nil {
				log.Printf("Failed to get err: %v", err)
			}
			if len(chats) == 0 {
				chatId = message.Chat.ID
				msg = fmt.Sprintf("Cannot find the target chat room\nPlease remind customer %s\n", orderInfo.AdvertiserUsername)
			} else {
				chatId = chats[0].ID
			}
			msg += fmt.Sprintf("Order %s canceled.", orderInfo.ID)

		case "已付款", "未付款", "争议中":
			pendingOrder := database.PendingOrder{
				MerchantOrderID:    orderInfo.MerchantOrderId,
				CustomerUsername:   orderInfo.CustomerUsername,
				AdvertiserUsername: orderInfo.AdvertiserUsername,
				OrderStatus:        orderInfo.OrderStatus,
				DisplayFiatAmount:  orderInfo.DisplayFiatAmount,
				Retries:            0,
				OriginalChatID:     message.Chat.ID,
				ReplyToMessageID:   int64(message.ReplyToMessage.ID),
			}
			if err := database.InsertPendingOrder(db, pendingOrder); err != nil {
				log.Printf("Failed to insert pending order: %v", err)
			}

			chats, err := database.GetChatByTitle(db, orderInfo.AdvertiserUsername)
			if err != nil {
				log.Printf("Failed to get err: %v", err)
			}
			if len(chats) == 0 {
				chatId = message.Chat.ID
				msg = fmt.Sprintf("Cannot find the target chat room\nPlease remind customer %s\n", orderInfo.AdvertiserUsername)
			} else {
				chatId = chats[0].ID
			}
			msg += fmt.Sprintf("Here are order %s need to confirm. Amount is %f Jcoin. Please check ASAP.", orderInfo.ID, orderInfo.DisplayFiatAmount)
		}

		if chatId == message.Chat.ID {
			// reply to the original message
			sendMessageParams := &bot.SendMessageParams{
				ChatID:          chatId,
				Text:            msg,
				ReplyParameters: &models.ReplyParameters{MessageID: message.ReplyToMessage.ID},
			}
			b.SendMessage(ctx, sendMessageParams)
		} else {
			if message.ReplyToMessage.Photo != nil && len(message.ReplyToMessage.Photo) > 0 {
				// Send photo with caption
				b.SendPhoto(ctx, &bot.SendPhotoParams{
					ChatID:  chatId,
					Photo:   &models.InputFileString{Data: message.ReplyToMessage.Photo[len(message.ReplyToMessage.Photo)-1].FileID},
					Caption: msg,
				})
				return
			}
			if message.ReplyToMessage.Document != nil {
				// Send document with caption
				b.SendDocument(ctx, &bot.SendDocumentParams{
					ChatID:   chatId,
					Document: &models.InputFileString{Data: message.ReplyToMessage.Document.FileID},
					Caption:  msg,
				})
				return
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   msg,
			})
		}
	}
}
