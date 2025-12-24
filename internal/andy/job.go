package andy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Largeb0525/personal-tool/database"
	"github.com/Largeb0525/personal-tool/internal/external/telegram"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

const (
	// DailyReportSchedule runs every day at 15:58
	DailyReportSchedule = "58 15 * * *"
	// UndelegateEnergySchedule runs every hour at the beginning of the hour
	// UndelegateEnergySchedule = "0 * * * *"
	// Vault2BotSchedule runs every hour at minute 55
	Vault2BotSchedule = "55 * * * *"
	// CheckPendingOrdersSchedule runs every minute
	CheckPendingOrdersSchedule = "* * * * *"
)

func StartCronJobs(ctx context.Context) *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc(DailyReportSchedule, scheduleDailyReport)
	if err != nil {
		panic(err)
	}

	// _, err = c.AddFunc(UndelegateEnergySchedule, undelegateEnergyJob)
	// if err != nil {
	// 	panic(err)
	// }

	_, err = c.AddFunc(Vault2BotSchedule, vault2BotJob)
	if err != nil {
		panic(err)
	}

	_, err = c.AddFunc(CheckPendingOrdersSchedule, checkPendingOrdersJob)
	if err != nil {
		panic(err)
	}

	c.Start()

	go func() {
		<-ctx.Done()
		c.Stop()
	}()
	return c
}

func scheduleDailyReport() {
	db := database.GetDB()
	countMap, err := database.GetTodayEventCountGroupByPlatform(db)
	if err != nil {
		log.Printf("Failed to get today event count: %v", err)
		return
	}

	delegatedCount, err := database.GetTodayDelegatedCount(db)
	if err != nil {
		log.Printf("Failed to get today delegated count: %v", err)
		return
	}

	message := fmt.Sprintf(
		`Daily Report
	%s: %d
	%s: %d
	Delegated: %d`,
		BName, countMap["b"], IName, countMap["i"], delegatedCount)

	err = telegram.SendTelegramMessage(message, telegram.TelegramChatId, telegram.TelegramBotToken)
	if err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
	}
}

func undelegateEnergyJob() {
	log.Printf("Undelegating energy...")
	db := database.GetDB()
	checkTime := time.Now().Add(-1 * time.Hour)
	delegateRecords, err := database.GetUndelegatedBefore(db, checkTime)
	if err != nil {
		log.Printf("Failed to get undelegated before: %v", err)
		return
	}

	for _, record := range delegateRecords {
		err := undelegateEnergy(record.ReceiverAddress, record.TxID)
		if err != nil {
			log.Printf("Failed to undelegate energy ,address: %s, txid: %s, err: %v", record.ReceiverAddress, record.TxID, err)
		}
	}
}

func vault2BotJob() {
	vault2Address := viper.GetString("andy.wallet.vault2_address")
	walletUsdt, err := getAddressUSDT(vault2Address)
	if err != nil {
		log.Printf("Failed to get address USDT: %v", err)
		return
	}
	walletUsdtFloat, _ := walletUsdt.Float64()
	message := fmt.Sprintf(
		`金庫2餘額：
	%f USDT`,
		walletUsdtFloat)

	err = telegram.SendTelegramMessage(message, "-829676684", telegram.TelegramVault2BotToken)
	if err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
	}
}

func checkPendingOrdersJob() {
	db := database.GetDB()
	orders, err := database.GetPendingOrders(db)
	if err != nil {
		log.Printf("Failed to get pending orders: %v", err)
		return
	}

	for _, order := range orders {
		if order.Retries >= 20 {
			log.Printf("Order %s reached max retries. Deleting.", order.MerchantOrderID)
			if err := database.DeletePendingOrder(db, order.MerchantOrderID); err != nil {
				log.Printf("Failed to delete pending order %s: %v", order.MerchantOrderID, err)
			}
			continue
		}

		latestOrderInfo, err := getIndiaOrder(order.MerchantOrderID, "merchant_order_id")
		if err != nil {
			log.Printf("Failed to get latest order info for %s: %v", order.MerchantOrderID, err)
			if err := database.IncrementPendingOrderRetries(db, order.MerchantOrderID); err != nil {
				log.Printf("Failed to increment retries for %s: %v", order.MerchantOrderID, err)
			}
			continue
		}

		msg := ""
		var chatId int64

		switch latestOrderInfo.OrderStatus {
		case "已完成":
			chats, err := database.GetChatByTitle(db, latestOrderInfo.CustomerUsername)
			if err != nil {
				log.Printf("Failed to get chat by title: %v", err)
			}
			if len(chats) == 0 {
				chatId = order.OriginalChatID
				msg = fmt.Sprintf("Cannot find the target chat room\nPlease remind customer %s\n", latestOrderInfo.AdvertiserUsername)
			} else {
				chatId = chats[0].ID
			}
			msg += fmt.Sprintf("Order %s completed.", latestOrderInfo.MerchantOrderId)

			// Check if the target chat is the same as the original chat
			if chatId == order.OriginalChatID {
				if err := telegram.SendReplyTelegramMessage(msg, fmt.Sprintf("%d", chatId), telegram.TelegramOrderBotToken, int(order.ReplyToMessageID)); err != nil {
					log.Printf("Failed to send reply Telegram message for completed order: %v", err)
				}
			} else {
				if err := telegram.SendTelegramMessage(msg, fmt.Sprintf("%d", chatId), telegram.TelegramOrderBotToken); err != nil {
					log.Printf("Failed to send Telegram message for completed order: %v", err)
				}
			}

			if err := database.DeletePendingOrder(db, order.MerchantOrderID); err != nil {
				log.Printf("Failed to delete completed pending order %s: %v", order.MerchantOrderID, err)
			}

		case "已取消":
			if err := database.DeletePendingOrder(db, order.MerchantOrderID); err != nil {
				log.Printf("Failed to delete canceled pending order %s: %v", order.MerchantOrderID, err)
			}

		case "已付款", "未付款", "争议中":
			if err := database.IncrementPendingOrderRetries(db, order.MerchantOrderID); err != nil {
				log.Printf("Failed to increment retries for %s: %v", order.MerchantOrderID, err)
			}
			continue
		}
	}
}
