package utils

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func SendError(c *gin.Context, code int, err string, msg string) {
	c.JSON(code, ErrorResponse{
		Error:   err,
		Message: msg,
	})
}

func SendSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}
