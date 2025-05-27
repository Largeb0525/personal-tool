package andy

import (
	"fmt"
	"log"
	"time"

	"github.com/Largeb0525/personal-tool/database"
	"github.com/Largeb0525/personal-tool/internal/external/telegram"
)

func ScheduleDailyReport() {
	location, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		panic(err)
	}

	for {
		now := time.Now().In(location)
		// 計算下一次 UTC+8 的 23:59
		next := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, location)
		if now.After(next) {
			// 若已過今日 23:59，則設為明日
			next = next.Add(24 * time.Hour)
		}
		duration := next.Sub(now)
		fmt.Printf("🕒 等待 %s 執行...\n", duration)
		time.Sleep(duration)

		db := database.GetDB()
		countMap, err := database.GetTodayEventCountGroupByPlatform(db)
		if err != nil {
			log.Printf("Failed to get today event count: %v", err)
			continue
		}

		message := fmt.Sprintf(
			`Daily Report
	%s: %d
	%s: %d`,
			BName, countMap["b"], IName, countMap["i"])

		err = telegram.SendTelegramMessage(message)
		if err != nil {
			log.Printf("Failed to send Telegram message: %v", err)
		}
	}
}
