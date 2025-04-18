package configs

import (
	"fmt"
	"net/http"

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
	//ctx := context.Background()

	var pipelines []PipelineDefinition
	if err := c.ShouldBindJSON(&pipelines); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	for _, pipeline := range pipelines {
		benthosCfgs, err := h.service.ParseJsonToBenthosConfig(pipeline)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}

		for nodeId, cfg := range benthosCfgs {
			fmt.Printf("nodeID: %s\n", nodeId)
			fmt.Println(string(cfg))
			fmt.Println("------------------------------------")

			// if err := h.dockerCli.LaunchBenthosContainer(ctx, nodeId, cfg); err != nil {
			// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			// 	return
			// }
		}
	}

	c.Status(http.StatusOK)
}
