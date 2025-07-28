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
	UndelegateEnergySchedule = "0 * * * *"
	// Vault2BotSchedule runs every hour at minute 55
	Vault2BotSchedule = "55 * * * *"
)

func StartCronJobs(ctx context.Context) *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc(DailyReportSchedule, scheduleDailyReport)
	if err != nil {
		panic(err)
	}

	_, err = c.AddFunc(UndelegateEnergySchedule, undelegateEnergyJob)
	if err != nil {
		panic(err)
	}

	_, err = c.AddFunc(Vault2BotSchedule, vault2BotJob)
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

	err = telegram.SendTelegramMessage(message, telegram.TelegramChatId, telegram.TelegramVault2BotToken)
	if err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
	}
}
