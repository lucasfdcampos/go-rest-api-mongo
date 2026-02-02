package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucas/go-rest-api-mongo/internal/config"
	"github.com/lucas/go-rest-api-mongo/internal/database"
	"github.com/lucas/go-rest-api-mongo/internal/handlers"
	"github.com/lucas/go-rest-api-mongo/internal/messaging"
	"github.com/lucas/go-rest-api-mongo/internal/middleware"
	"github.com/lucas/go-rest-api-mongo/internal/repositories"
	"github.com/lucas/go-rest-api-mongo/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	db, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			log.Printf("Error closing database: %v\n", err)
		}
	}()

	log.Println("✅ Connected to MongoDB")

	kafkaProducer := messaging.NewKafkaProducer(cfg)
	defer kafkaProducer.Close()

	log.Println("✅ Kafka Producer initialized")

	userRepository := repositories.NewUserRepository(db)
	authService := services.NewAuthService(cfg)
	userService := services.NewUserService(userRepository, authService)

	workerPool := services.NewWorkerPool(
		userService,
		kafkaProducer,
		cfg.Workers.PoolSize,
		cfg.Workers.BatchSize,
		cfg.Workers.BatchTimeout,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerPool.Start(ctx)
	log.Printf("✅ Worker Pool started with %d workers\n", cfg.Workers.PoolSize)

	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(workerPool)

	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	setupRoutes(router, authHandler, userHandler, authService)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	go func() {
		log.Printf("🚀 Server starting on %s:%s\n", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
	}

	log.Println("✅ Server stopped gracefully")
}

func setupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, authService *services.AuthService) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Rotas públicas (sem autenticação)
	public := router.Group("/api/v1")
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
		public.POST("/register-fast", userHandler.Register) // Assíncrono com Worker Pool
	}

	// Rotas protegidas (com autenticação)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID := c.GetString("user_id")
			email := c.GetString("email")
			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"email":   email,
				"message": "This is a protected route",
			})
		})
	}

	log.Println("✅ Routes configured")
}
