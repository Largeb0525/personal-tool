package internal

import (
	"net/http"

	"github.com/Largeb0525/personal-tool/andy"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	andy.AndyRouter(r)

	return r
}
