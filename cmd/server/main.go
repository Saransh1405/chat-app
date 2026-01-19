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

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(cfg.Database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	wsHub := websocket.NewHub()
	go wsHub.Run()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	kafka, err := kafka.NewKafka(cfg.Kafka.Brokers)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer kafka.Close()

	go func() {
		err = kafka.Produce(cfg.Kafka.Topic, []byte("test message"))
		if err != nil {
			log.Fatalf("Failed to produce message: %v", err)
		}

		err = kafka.StartConsumer(cfg.Kafka.Topic, 0, int64(sarama.OffsetNewest))
		if err != nil {
			log.Fatalf("Failed to start Kafka consumer: %v", err)
		}
	}()

	redisClient, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	defer redisClient.Client.Close()

	router := gin.New()

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(cfg.CORS))

	router.GET("/health", healthCheck)
	router.GET("/health/ready", readinessCheck(db))
	router.GET("/health/live", livenessCheck)

	v1 := router.Group("/api/v1")
	{
		authHandler := rest.NewAuthHandler(cfg, db)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		protected := v1.Group("")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			appHandler := rest.NewApplicationHandler(db)
			apps := protected.Group("/applications")
			{
				apps.POST("", appHandler.Create)
				apps.GET("", appHandler.List)
				apps.GET("/:id", appHandler.Get)
				apps.PUT("/:id", appHandler.Update)
				apps.DELETE("/:id", appHandler.Delete)
			}

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

			messageHandler := rest.NewMessageHandler(db, wsHub)
			messages := protected.Group("/applications/:app_id/rooms/:room_id/messages")
			{
				messages.Use(middleware.Auth(cfg.JWT.Secret))
				messages.POST("", messageHandler.Create)
				messages.GET("", messageHandler.List)
				messages.GET("/:message_id", messageHandler.Get)
				messages.PUT("/:message_id", messageHandler.Update)
				messages.DELETE("/:message_id", messageHandler.Delete)
			}

			reactionHandler := rest.NewReactionHandler(db, wsHub)
			reactions := protected.Group("/applications/:app_id/messages/:message_id/reactions")
			{
				reactions.POST("", reactionHandler.Create)
				reactions.DELETE("/:reaction_id", reactionHandler.Delete)
				reactions.GET("", reactionHandler.List)
			}

			typingHandler := rest.NewTypingHandler(db, wsHub)
			typing := protected.Group("/applications/:app_id/rooms/:room_id/typing")
			{
				typing.POST("", typingHandler.Create)
			}
		}

		wsHandler := websocket.NewHandler(wsHub, cfg, db)
		ws := protected.Group("/ws")
		{
			ws.Use(middleware.Auth(cfg.JWT.Secret))
			ws.GET("", wsHandler.HandleConnection)
		}
	}

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

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
