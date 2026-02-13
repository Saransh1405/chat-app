package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat-app/internal/config"
	"chat-app/internal/database"
	"chat-app/internal/handler/rest"
	"chat-app/internal/handler/websocket"
	imagekit "chat-app/internal/image-kit"
	"chat-app/internal/middleware"
	"chat-app/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", err, logger.Fields{
			"error": err.Error(),
		})
	}

	// Initialize logger
	logger.Initialize(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("Logger initialized", logger.Fields{
		"level":  cfg.Logging.Level,
		"format": cfg.Logging.Format,
	})

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", err, logger.Fields{
			"host": cfg.Database.Host,
			"port": cfg.Database.Port,
			"name": cfg.Database.Name,
		})
	}
	defer db.Close()

	logger.Info("Database connection established", logger.Fields{
		"host": cfg.Database.Host,
		"port": cfg.Database.Port,
		"name": cfg.Database.Name,
	})

	if cfg.Environment == "DEVELOPMENT" {
		if err := database.RunMigrations(cfg.Database); err != nil {
			logger.Fatal("Failed to run migrations", err, logger.Fields{})
		}
	}

	logger.Info("Database migrations completed successfully")

	wsHub := websocket.NewHub()
	go wsHub.Run()

	if cfg.Environment == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	ik, err := imagekit.InitImageKit(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize ImageKit client", err, logger.Fields{})
	}

	// kafka, err := kafka.NewKafka(cfg.Kafka.Brokers)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize Kafka producer: %v", err)
	// }
	// defer kafka.Close()

	// go func() {
	// 	err = kafka.Produce(cfg.Kafka.Topic, []byte("test message"))
	// 	if err != nil {
	// 		log.Fatalf("Failed to produce message: %v", err)
	// 	}

	// 	err = kafka.StartConsumer(cfg.Kafka.Topic, 0, int64(sarama.OffsetNewest))
	// 	if err != nil {
	// 		log.Fatalf("Failed to start Kafka consumer: %v", err)
	// 	}
	// }()

	// redisClient, err := redis.NewRedis(cfg.Redis)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize Redis client: %v", err)
	// }
	// defer redisClient.Client.Close()

	router := gin.New()

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(cfg.CORS))

	router.GET("/health", healthCheck)
	router.GET("/health/ready", readinessCheck(db))
	router.GET("/health/live", livenessCheck)

	fileUploadHandler := rest.NewFileUploadHandler(db, ik)
	fileUpload := router.Group("/api/v1/file/upload")
	{
		fileUpload.POST("", fileUploadHandler.Upload)
	}

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
				apps.GET("", appHandler.Get)
				apps.PATCH("", appHandler.Update)
				apps.DELETE("", appHandler.Delete)
			}

			userHandler := rest.NewUserHandler(db)
			users := protected.Group("/users")
			{
				users.POST("", userHandler.Create)
				users.GET("", userHandler.Get)
				users.PATCH("", userHandler.Update)
				users.DELETE("", userHandler.Delete)
				users.GET("/all", userHandler.GetUsers)
			}

			roomHandler := rest.NewRoomHandler(db, wsHub)
			rooms := protected.Group("/rooms")
			{
				rooms.POST("", roomHandler.Create)
				rooms.GET("", roomHandler.List)
				rooms.PATCH("", roomHandler.Update)
				rooms.DELETE("", roomHandler.Delete)
				rooms.POST("/members", roomHandler.AddMember)
				rooms.DELETE("/members", roomHandler.RemoveMember)
				rooms.GET("/members", roomHandler.ListMembers)
			}

			messageHandler := rest.NewMessageHandler(db, wsHub)
			messages := protected.Group("/messages")
			{
				messages.POST("", messageHandler.Create)
				messages.GET("", messageHandler.List)
				messages.PATCH("", messageHandler.Update)
				messages.DELETE("", messageHandler.Delete)
			}

			reactionHandler := rest.NewReactionHandler(db, wsHub)
			reactions := protected.Group("/reactions")
			{
				reactions.POST("", reactionHandler.Create)
				reactions.DELETE("", reactionHandler.Delete)
				reactions.GET("", reactionHandler.List)
			}

			typingHandler := rest.NewTypingHandler(db, wsHub)
			typing := protected.Group("/typing")
			{
				typing.POST("", typingHandler.Create)
			}
		}

		wsHandler := websocket.NewHandler(wsHub, cfg, db)
		ws := v1.Group("/ws")
		{
			// WebSocket handler handles authentication internally (supports token in query param)
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

	logger.Info("ImageKit client initialized", logger.Fields{})

	go func() {
		logger.Info("Server starting", logger.Fields{
			"address": addr,
			"host":    cfg.Server.Host,
			"port":    cfg.Server.Port,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err, logger.Fields{
				"address": addr,
			})
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...", logger.Fields{})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err, logger.Fields{})
	}

	logger.Info("Server exited gracefully", logger.Fields{})
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
