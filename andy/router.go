package andy

import (
	"github.com/gin-gonic/gin"
)

func AndyRouter(r *gin.Engine) {
	routerGroup := r.Group("/andy")
	{
		routerGroup.POST("/request", handlePostRequest)
		routerGroup.POST("/event", eventHandler)
	}

}
