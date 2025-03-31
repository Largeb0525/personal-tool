package andy

import (
	"fmt"
	"log"
	"time"
)

func ScheduleDailyReport() {
	location, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		panic(err)
	}

	for {
		now := time.Now().In(location)
		// 計算下一次 UTC+8 的 00:00
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, location)
		duration := next.Sub(now)

		fmt.Printf("🕒 等待 %s 執行...\n", duration)
		time.Sleep(duration)

		// 執行你的任務
		message := fmt.Sprintf(
			`Daily Report
	%s: %d
	%s: %d`,
			BName, BDailyCount, IName, IDailyCount)

		payload := TelegramMessagePayload{
			ChatID: TelegramChatId,
			Text:   message,
		}
		err = SendTelegramMessage(payload)
		if err != nil {
			log.Printf("Failed to send Telegram message: %v", err)
		}
		BDailyCount = 0
		IDailyCount = 0
	}
}
