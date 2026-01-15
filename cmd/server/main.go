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

	"chat-app/internal/config"
	"chat-app/internal/database"
	"chat-app/internal/handler/rest"
	"chat-app/internal/handler/websocket"
	"chat-app/internal/kafka"
	"chat-app/internal/middleware"
	"chat-app/internal/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(cfg.Database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Kafka producer
	kafka, err := kafka.NewKafka(cfg.Kafka.Brokers)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer kafka.Close()

	// // Initialize Kafka consumer
	// kafkaConsumer, err := kafka.NewKafkaConsumer(cfg.Kafka.Brokers)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	// }
	// defer kafkaConsumer.Close()

	// Initialize redis client
	redisClient, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	defer redisClient.Client.Close()

	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(cfg.CORS))

	// Health check endpoints
	router.GET("/health", healthCheck)
	router.GET("/health/ready", readinessCheck(db))
	router.GET("/health/live", livenessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (authentication)
		authHandler := rest.NewAuthHandler(cfg, db)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Application routes
			appHandler := rest.NewApplicationHandler(db)
			apps := protected.Group("/applications")
			{
				apps.POST("", appHandler.Create)
				apps.GET("", appHandler.List)
				apps.GET("/:id", appHandler.Get)
				apps.PUT("/:id", appHandler.Update)
				apps.DELETE("/:id", appHandler.Delete)
			}

			// User routes (scoped to application)
			userHandler := rest.NewUserHandler(db)
			users := protected.Group("/applications/:app_id/users")
			{
				users.POST("", userHandler.Create)
			}
			users.Use(middleware.Auth(cfg.JWT.Secret))
			{
				users.GET("", userHandler.Get)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			// User routes
			directUserHandler := rest.NewUserHandler(db)
			directUsers := protected.Group("/users")
			{
				directUsers.POST("", directUserHandler.Create)
			}
			directUsers.Use(middleware.Auth(cfg.JWT.Secret))
			{
				directUsers.GET("", directUserHandler.Get)
				directUsers.PUT("/:id", directUserHandler.Update)
				directUsers.DELETE("/:id", directUserHandler.Delete)
			}

			// Room routes (scoped to application)
			roomHandler := rest.NewRoomHandler(db)
			rooms := protected.Group("/applications/:app_id/rooms")
			{
				rooms.POST("", roomHandler.Create)
				rooms.GET("", roomHandler.List)
				rooms.GET("/:id", roomHandler.Get)
				rooms.PUT("/:id", roomHandler.Update)
				rooms.DELETE("/:id", roomHandler.Delete)
				rooms.POST("/:id/members", roomHandler.AddMember)
				rooms.DELETE("/:id/members/:user_id", roomHandler.RemoveMember)
				rooms.GET("/:id/members", roomHandler.ListMembers)
			}

			// Message routes (scoped to application and room)
			messageHandler := rest.NewMessageHandler(db, wsHub)
			messages := protected.Group("/applications/:app_id/rooms/:room_id/messages")
			{
				messages.POST("", messageHandler.Create)
				messages.GET("", messageHandler.List)
				messages.GET("/:message_id", messageHandler.Get)
				messages.PUT("/:message_id", messageHandler.Update)
				messages.DELETE("/:message_id", messageHandler.Delete)
			}

			// Reaction routes
			reactionHandler := rest.NewReactionHandler(db, wsHub)
			reactions := protected.Group("/applications/:app_id/messages/:message_id/reactions")
			{
				reactions.POST("", reactionHandler.Create)
				reactions.DELETE("/:reaction_id", reactionHandler.Delete)
				reactions.GET("", reactionHandler.List)
			}

			// Typing indicator routes
			typingHandler := rest.NewTypingHandler(db, wsHub)
			typing := protected.Group("/applications/:app_id/rooms/:room_id/typing")
			{
				typing.POST("", typingHandler.Create)
			}
		}

		// WebSocket endpoint
		wsHandler := websocket.NewHandler(wsHub, cfg, db)
		ws := protected.Group("/ws")
		{
			ws.Use(middleware.Auth(cfg.JWT.Secret))
			ws.GET("", wsHandler.HandleConnection)
		}
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func readinessCheck(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "database connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC(),
		})
	}
}

func livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
	})
}
