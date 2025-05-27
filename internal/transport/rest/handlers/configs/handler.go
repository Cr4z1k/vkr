package configs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Cr4z1k/vkr/internal/model"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   Service
	dockerCli Docker
}

func New(service Service, dockerCli Docker) *Handler {
	return &Handler{
		service:   service,
		dockerCli: dockerCli,
	}
}

func (h *Handler) SetConfigs(c *gin.Context) {
	ctx := context.Background()

	var pipelines []model.PipelineDefinition
	if err := c.ShouldBindJSON(&pipelines); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})

		log.Printf("Invalid JSON payload: %s", err.Error())

		return
	}

	if err := h.dockerCli.CleanupRemovedContainers(ctx, pipelines); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error CleanupRemovedContainers: %s", err.Error())})

		log.Printf("Error cleaning up removed containers: %s", err.Error())

		return
	}

	log.Printf("Cleaned up containers")

	for _, pipeline := range pipelines {
		benthosCfgPaths, err := h.service.ParseJsonToBenthosConfig(pipeline)
		if errors.Is(err, model.ErrBenthosValidation) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":           "config validation error",
				"validation info": err.Error(),
			})

			log.Printf("Config validation error: %s", err.Error())

			return
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server error"})

			log.Printf("Error parsing JSON to Redpanda Connect config: %s", err.Error())

			return
		}

		for nodeId, cfgPath := range benthosCfgPaths {
			if err := h.dockerCli.LaunchBenthosContainer(ctx, pipeline.Name, nodeId, cfgPath); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error LaunchBenthosContainer: %s", err.Error())})

				log.Printf("Error launching Redpanda Connect container for pipeline %s, node %s: %s", pipeline.Name, nodeId, err.Error())

				return
			}
		}
	}

	log.Printf("Successfully launched Redpanda Connect containers for pipelines")

	c.Status(http.StatusOK)
}
