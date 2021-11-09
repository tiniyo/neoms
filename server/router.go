package server

import (
	"github.com/tiniyo/neoms/controller"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	callController := new(controller.CallController)

	callController.InitializeCallController()

	rootPath := router.Group("/api")
	{
		versionPath := rootPath.Group("v1")
		{
			account := versionPath.Group("account")
			{
				account.POST(":account_id/Call/:call_id", callController.CreateCall)
				account.PUT(":account_id/Call/:call_id", callController.UpdateCall)
				account.GET(":account_id/Call/:call_id", callController.GetCall)
				account.DELETE(":account_id/Call/:call_id", callController.DeleteCall)
			}

			versionPath.GET("health", callController.GetHealth)
		}
	}
	return router
}
