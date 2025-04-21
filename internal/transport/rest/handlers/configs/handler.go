package configs

import (
	"context"
	"errors"
	"fmt"
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

	var pipelines []PipelineDefinition
	if err := c.ShouldBindJSON(&pipelines); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	if err := h.dockerCli.CleanupRemovedContainers(ctx, pipelines); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error CleanupRemovedContainers: %s", err.Error())})
		return
	}

	for _, pipeline := range pipelines {
		benthosCfgPaths, err := h.service.ParseJsonToBenthosConfig(pipeline)
		if errors.Is(err, model.ErrBenthosValidation) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":           "config validation error",
				"validation info": err.Error(),
			})

			return
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error ParseJsonToBenthosConfig: %s", err.Error())})
			return
		}

		for nodeId, cfgPath := range benthosCfgPaths {
			if err := h.dockerCli.LaunchBenthosContainer(ctx, pipeline.Name, nodeId, cfgPath); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error LaunchBenthosContainer: %s", err.Error())})
				return
			}
		}
	}

	c.Status(http.StatusOK)
}
