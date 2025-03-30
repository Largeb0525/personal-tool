package internal

import (
	"net/http"

	"github.com/Largeb0525/personal-tool/andy"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	fillParameters()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	andy.AndyRouter(r)

	return r
}

func fillParameters() {
	andy.BName = viper.GetString("andy.b.name")
	andy.IName = viper.GetString("andy.i.name")
	andy.TelegramChatId = viper.GetString("andy.telegram.chat_id")
	andy.EnergyToken = viper.GetString("andy.energy.token")
	andy.EnergyUrl = viper.GetString("andy.energy.url")
	andy.TelegramBotToken = viper.GetString("andy.telegram.bot_token")
}
