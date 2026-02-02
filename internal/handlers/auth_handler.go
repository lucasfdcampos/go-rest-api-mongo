package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucas/go-rest-api-mongo/internal/dto"
	"github.com/lucas/go-rest-api-mongo/internal/services"
	"github.com/lucas/go-rest-api-mongo/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		// Trata erros específicos
		if errors.Is(err, services.ErrEmailExists) {
			utils.SendError(c, http.StatusConflict, "conflict", "email already exists")
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "internal_error", "failed to register user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user": dto.UserResponse{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.String(),
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	loginResponse, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		// Trata erros específicos
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials")
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "internal_error", "failed to login")
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", "user not authenticated")
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "bad_request", "invalid user ID")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), objectID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			utils.SendError(c, http.StatusNotFound, "not_found", "user not found")
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "internal_error", "failed to retrieve user profile")
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
	})
}
