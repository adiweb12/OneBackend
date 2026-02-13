package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"onechat/internal/config"
	"onechat/internal/database"
	"onechat/internal/handlers"
	"onechat/internal/middleware"
	"onechat/internal/services"
	"onechat/internal/websocket"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Auto-migrate database schemas
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(db, cfg.JWTSecret)
	chatService := services.NewChatService(db)
	groupService := services.NewGroupService(db)
	aiService := services.NewAIService(cfg.GeminiAPIKey)
	mediaService := services.NewMediaService(cfg.CloudinaryURL)
	eventService := services.NewEventService(db, aiService)
	notificationService := services.NewNotificationService()

	// Initialize WebSocket hub
	hub := websocket.NewHub(chatService)
	go hub.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	chatHandler := handlers.NewChatHandler(chatService, hub)
	groupHandler := handlers.NewGroupHandler(groupService, hub)
	aiHandler := handlers.NewAIHandler(aiService)
	mediaHandler := handlers.NewMediaHandler(mediaService)
	eventHandler := handlers.NewEventHandler(eventService)
	wsHandler := handlers.NewWebSocketHandler(hub, authService)

	// Setup router
	router := setupRouter(cfg, authHandler, chatHandler, groupHandler, aiHandler, mediaHandler, eventHandler, wsHandler)

	// Start media cleanup scheduler
	go mediaService.StartCleanupScheduler(10 * 24 * time.Hour) // 10 days

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	chatHandler *handlers.ChatHandler,
	groupHandler *handlers.GroupHandler,
	aiHandler *handlers.AIHandler,
	mediaHandler *handlers.MediaHandler,
	eventHandler *handlers.EventHandler,
	wsHandler *handlers.WebSocketHandler,
) *gin.Engine {
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", authHandler.GetProfile)
				users.PUT("/me", authHandler.UpdateProfile)
				users.GET("/search", authHandler.SearchUsers)
			}

			// Chat routes
			chats := protected.Group("/chats")
			{
				chats.GET("", chatHandler.GetChats)
				chats.POST("", chatHandler.CreateChat)
				chats.GET("/:chatId/messages", chatHandler.GetMessages)
				chats.POST("/:chatId/messages", chatHandler.SendMessage)
				chats.PUT("/messages/:messageId/status", chatHandler.UpdateMessageStatus)
				chats.DELETE("/messages/:messageId", chatHandler.DeleteMessage)
			}

			// Group routes
			groups := protected.Group("/groups")
			{
				groups.POST("", groupHandler.CreateGroup)
				groups.GET("/:groupId", groupHandler.GetGroup)
				groups.PUT("/:groupId", groupHandler.UpdateGroup)
				groups.DELETE("/:groupId", groupHandler.DeleteGroup)
				groups.POST("/:groupId/members", groupHandler.AddMember)
				groups.DELETE("/:groupId/members/:userId", groupHandler.RemoveMember)
				groups.PUT("/:groupId/members/:userId/role", groupHandler.UpdateMemberRole)
			}

			// AI routes
			ai := protected.Group("/ai")
			{
				ai.POST("/research", aiHandler.Research)
				ai.POST("/extract-event", aiHandler.ExtractEvent)
			}

			// Media routes
			media := protected.Group("/media")
			{
				media.POST("/upload", mediaHandler.Upload)
			}

			// Event routes
			events := protected.Group("/events")
			{
				events.GET("", eventHandler.GetEvents)
				events.POST("", eventHandler.CreateEvent)
				events.PUT("/:eventId", eventHandler.UpdateEvent)
				events.DELETE("/:eventId", eventHandler.DeleteEvent)
			}
		}
	}

	// WebSocket route
	router.GET("/ws", middleware.WSAuthMiddleware(cfg.JWTSecret), wsHandler.HandleWebSocket)

	return router
}
