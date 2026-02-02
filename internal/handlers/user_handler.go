package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucas/go-rest-api-mongo/internal/dto"
	"github.com/lucas/go-rest-api-mongo/internal/services"
	"github.com/lucas/go-rest-api-mongo/pkg/utils"
)

type UserHandler struct {
	workerPool *services.WorkerPool
}

func NewUserHandler(workerPool *services.WorkerPool) *UserHandler {
	return &UserHandler{
		workerPool: workerPool,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	if err := h.workerPool.Submit(&req); err != nil {
		utils.SendError(c, http.StatusServiceUnavailable, "service_unavailable", "too many requests")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "registration request accepted",
	})
}
