package internal

import (
	"net/http"
	"strings"

	"github.com/Largeb0525/personal-tool/internal/andy"
	quicknode "github.com/Largeb0525/personal-tool/internal/external/quickNode"
	"github.com/Largeb0525/personal-tool/internal/external/telegram"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	fillParameters()
	go andy.ScheduleDailyReport()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/reFillParameters", func(c *gin.Context) {
		fillParameters()
		c.JSON(http.StatusOK, gin.H{
			"andy_b_name":        maskStart(andy.BName),
			"andy_i_name":        maskStart(andy.IName),
			"energy_token":       maskStart(andy.EnergyToken),
			"energy_url":         maskStart(andy.EnergyUrl),
			"telegram_chat_id":   maskStart(telegram.TelegramChatId),
			"telegram_bot_token": maskStart(telegram.TelegramBotToken),
			"quicknode_api_key":  maskStart(quicknode.ApiKey),
			"quicknode_alert_id": maskStart(quicknode.QuickAlertID),
		})
	})

	andy.AndyRouter(r)

	return r
}

func fillParameters() {
	andy.BName = viper.GetString("andy.b.name")
	andy.IName = viper.GetString("andy.i.name")
	andy.EnergyToken = viper.GetString("andy.energy.token")
	andy.EnergyUrl = viper.GetString("andy.energy.url")
	telegram.TelegramChatId = viper.GetString("andy.telegram.chat_id")
	telegram.TelegramBotToken = viper.GetString("andy.telegram.bot_token")
	quicknode.ApiKey = viper.GetString("quicknode.api_key")
	quicknode.QuickAlertID = viper.GetString("quicknode.quick_alert_id")
}

func maskStart(s string) string {
	if len(s) <= 5 {
		return strings.Repeat("*", len(s))
	}
	return strings.Repeat("*", 5) + s[5:]
}
