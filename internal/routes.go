package internal

import (
	"context"
	"net/http"
	"strings"

	"github.com/Largeb0525/personal-tool/internal/andy"
	"github.com/Largeb0525/personal-tool/internal/external/quickNode"
	"github.com/Largeb0525/personal-tool/internal/external/telegram"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

func InitRouter(ctx context.Context) (*gin.Engine, *cron.Cron) {
	r := gin.Default()
	fillParameters()
	c := andy.StartCronJobs(ctx)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/reFillParameters", func(c *gin.Context) {
		fillParameters()
		c.JSON(http.StatusOK, gin.H{
			"andy_b_name":               maskStart(andy.BName),
			"andy_i_name":               maskStart(andy.IName),
			"energy_token":              maskStart(andy.EnergyToken),
			"energy_url":                maskStart(andy.EnergyUrl),
			"energy_address":            maskStart(andy.EnergyAddress),
			"tron_private_key":          maskStart(andy.TronPrivateKey),
			"telegram_chat_id":          maskStart(telegram.TelegramChatId),
			"telegram_critical_chat_id": maskStart(telegram.CriticalTelegramChatId),
			"telegram_bot_token":        maskStart(telegram.TelegramBotToken),
			"telegram_vault2_bot_token": maskStart(telegram.TelegramVault2BotToken),
			"quicknode_api_key":         maskStart(quickNode.ApiKey),
			"quicknode_app_id":          maskStart(quickNode.AppID),
			"quicknode_alert_id":        maskStart(quickNode.QuickAlertID),
		})
	})

	andy.AndyRouter(r)

	return r, c
}

func fillParameters() {
	andy.BName = viper.GetString("andy.b.name")
	andy.IName = viper.GetString("andy.i.name")
	andy.EnergyToken = viper.GetString("andy.energy.token")
	andy.EnergyUrl = viper.GetString("andy.energy.url")
	andy.EnergyAddress = viper.GetString("tron.energy_address")
	andy.TronPrivateKey = viper.GetString("tron.private_key")
	telegram.TelegramChatId = viper.GetString("andy.telegram.chat_id")
	telegram.CriticalTelegramChatId = viper.GetString("andy.telegram.critical_chat_id")
	telegram.TelegramBotToken = viper.GetString("andy.telegram.bot_token")
	telegram.TelegramVault2BotToken = viper.GetString("andy.telegram.vault2_bot_token")
	quickNode.ApiKey = viper.GetString("quicknode.api_key")
	quickNode.AppID = viper.GetString("quicknode.app_id")
	quickNode.QuickAlertID = viper.GetString("quicknode.quick_alert_id")
}

func maskStart(s string) string {
	if len(s) <= 5 {
		return strings.Repeat("*", len(s))
	}
	return strings.Repeat("*", 5) + s[5:]
}
