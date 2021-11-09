package server

import (
	"github.com/gin-gonic/gin"
	"github.com/tiniyo/neoms/config"
)

func Init() {
	gin.SetMode(config.Config.Server.GinMode)
	r := NewRouter()
	r.Run(config.Config.Server.Port)
}
