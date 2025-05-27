package clean_up

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	dockerCli Docker
}

func New(dockerCli Docker) *Handler {
	return &Handler{
		dockerCli: dockerCli,
	}
}

func (h *Handler) CleanUp(c *gin.Context) {
	ctx := context.Background()

	if err := h.dockerCli.CleanupPipelinesContainers(ctx); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "server error",
		})

		log.Printf("Error cleaning up pipelines Docker containers: %s", err.Error())

		return
	}

	log.Printf("Cleaned up pipelines Docker containers")

	c.Status(http.StatusOK)
}
