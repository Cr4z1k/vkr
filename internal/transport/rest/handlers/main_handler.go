package handlers

import (
	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers/configs"
	"github.com/gin-gonic/gin"
)

type MainHandler struct {
	configsHandler *configs.Handler
}

func New(configsHandler *configs.Handler) *MainHandler {
	return &MainHandler{
		configsHandler: configsHandler,
	}
}

func (h *MainHandler) InitRoutes() *gin.Engine {
	r := gin.New()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/setConfigs", h.configsHandler.SetConfigs)

	return r
}
