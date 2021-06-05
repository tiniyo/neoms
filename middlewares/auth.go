package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, password, ok := c.Request.BasicAuth()

		if ok && user == "admin" && password == "admin" {
			c.Next()
		} else {
			response := gin.H{
				"status":  "error",
				"message": "Not Authorized",
			}
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
		}
	}
}
