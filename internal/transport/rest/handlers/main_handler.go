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

	r.POST("/setConfigs", h.configsHandler.SetConfigs)

	return r
}
