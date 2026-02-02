package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lucas/go-rest-api-mongo/internal/services"
	"github.com/lucas/go-rest-api-mongo/pkg/utils"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Pega o header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized", "missing authorization header")
			c.Abort()
			return
		}

		// 2. Verifica se começa com "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized", "invalid authorization format")
			c.Abort()
			return
		}

		// 3. Extrai o token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 4. Valida o token
		token, err := authService.ValidateToken(tokenString)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized", "invalid token")
			c.Abort()
			return
		}

		// 5. Extrai as claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized", "invalid token claims")
			c.Abort()
			return
		}

		// 6. Guarda user_id e email no context para uso nos handlers
		if userID, ok := claims["user_id"].(string); ok {
			c.Set("user_id", userID)
		}
		if email, ok := claims["email"].(string); ok {
			c.Set("email", email)
		}

		// 7. Continua para o próximo handler
		c.Next()
	}
}
