package andy

import (
	"github.com/gin-gonic/gin"
)

func AndyRouter(r *gin.Engine) {
	routerGroup := r.Group("/andy")
	{
		routerGroup.POST("/request", handleOrderRequestHandler)
		routerGroup.POST("/event/:platform", quickAlertsEventHandler)
		routerGroup.PATCH("/event/threshold", thresholdHandler)
		routerGroup.GET("/event/daily-report", dailyReportHandler)
		routerGroup.POST("/refresh/:platform", refreshHandler)
		routerGroup.POST("/upload/address/:platform", uploadAddressCsvFileHandler)
		routerGroup.POST("/freezeTRX", freezeHandler)
	}
}
