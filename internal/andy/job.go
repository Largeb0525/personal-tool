package andy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Largeb0525/personal-tool/database"
	"github.com/Largeb0525/personal-tool/internal/external/telegram"

	"github.com/robfig/cron/v3"
)

func StartCronJobs(ctx context.Context) *cron.Cron {
	c := cron.New()

	// per day
	_, err := c.AddFunc("58 15 * * *", scheduleDailyReport)
	if err != nil {
		panic(err)
	}

	// per hour
	_, err = c.AddFunc("0 * * * *", undelegateEnergyJob)
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

	err = telegram.SendTelegramMessage(message)
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
