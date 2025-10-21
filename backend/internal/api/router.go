package api

import (
	"database/sql"
	
	"github.com/gin-gonic/gin"
	"ai-doc-system/internal/auth"
	"ai-doc-system/internal/services"
)

func SetupRouter(db *sql.DB, jwtSecret string) *gin.Engine {
	r := gin.Default()
	
	// Set file upload size limit
	r.MaxMultipartMemory = 10 << 20 // 10MB
	
	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Service is healthy",
		})
	})
	
	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// Initialize services and handlers
	userService := services.NewUserService(db)
	userHandler := NewUserHandler(userService, jwtSecret)
	
	fileService := services.NewFileService(db, "storage/files")
	fileHandler := NewFileHandler(fileService, jwtSecret)
	
	friendService := services.NewFriendService(db)
	friendHandler := NewFriendHandler(friendService)
	
	messageService := services.NewMessageService(db)
	messageHandler := NewMessageHandler(messageService)
	
	fileShareService := services.NewFileShareService(db)
	fileShareHandler := NewFileShareHandler(fileShareService, fileService)
	
	onlyOfficeHandler := NewOnlyOfficeHandler(jwtSecret, fileService)
	
	// User authentication routes (no authentication required)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", userHandler.Register)
		authGroup.POST("/login", userHandler.Login)
	}
	
	// Public shared file download (no authentication required)
	r.GET("/api/share/:token", fileShareHandler.DownloadSharedFile)
	
	// File edit, preview and download endpoints with their own auth logic for iframe support
	r.GET("/api/files/:id/edit", fileHandler.EditFile)
	r.GET("/api/files/:id/preview", fileHandler.PreviewFile)
	r.GET("/api/files/:id/download", fileHandler.DownloadFile)
	
	// OnlyOffice integration endpoints
	r.GET("/api/onlyoffice/config/:id", onlyOfficeHandler.GetOnlyOfficeConfig)
	r.GET("/api/files/:id/onlyoffice/config", onlyOfficeHandler.GetOnlyOfficeConfig)
	r.POST("/api/onlyoffice/callback", onlyOfficeHandler.HandleCallback)
	
	// Protected routes (authentication required)
	protected := r.Group("/api")
	protected.Use(auth.AuthMiddleware(jwtSecret))
	{
		// User related
		protected.GET("/profile", userHandler.GetProfile)
		protected.PUT("/profile", userHandler.UpdateProfile)
		
		// File related
		protected.POST("/files/upload", fileHandler.UploadFile)
		protected.GET("/files", fileHandler.GetUserFiles)
		protected.GET("/files/:id", fileHandler.GetFile)
		protected.DELETE("/files/:id", fileHandler.DeleteFile)
		protected.PUT("/files/:id/rename", fileHandler.RenameFile)
		protected.GET("/storage/usage", fileHandler.GetStorageUsage)
		
		// Friend related
		protected.POST("/friends/request", friendHandler.SendFriendRequest)
		protected.POST("/friends/accept/:id", friendHandler.AcceptFriendRequest)
		protected.POST("/friends/reject/:id", friendHandler.RejectFriendRequest)
		protected.DELETE("/friends/:id", friendHandler.RemoveFriend)
		protected.GET("/friends", friendHandler.GetFriends)
		protected.GET("/friends/requests", friendHandler.GetPendingRequests)
		protected.GET("/users/search", friendHandler.SearchUsers)
		
		// Friend groups
		protected.POST("/friend-groups", friendHandler.CreateFriendGroup)
		protected.GET("/friend-groups", friendHandler.GetFriendGroups)
		protected.POST("/friend-groups/:id/add-friend", friendHandler.AddFriendToGroup)
		
		// Message related
		protected.POST("/messages", messageHandler.SendMessage)
		protected.GET("/messages/:friend_id", messageHandler.GetChatHistory)
		protected.GET("/chats", messageHandler.GetChatList)
		protected.PUT("/messages/:id/read", messageHandler.MarkAsRead)
		protected.GET("/messages/unread/count", messageHandler.GetUnreadCount)
		protected.DELETE("/messages/:id", messageHandler.DeleteMessage)
		
		// File sharing
		protected.POST("/shares/friend", fileShareHandler.ShareToFriend)
		protected.POST("/shares/public", fileShareHandler.CreatePublicShare)
		protected.GET("/shares/with-me", fileShareHandler.GetSharedWithMe)
		protected.GET("/shares/my-shares", fileShareHandler.GetMyShares)
		protected.GET("/shares/files/:id/download", fileShareHandler.DownloadFriendSharedFile)
		protected.DELETE("/shares/:id", fileShareHandler.RemoveShare)
	}
	
	// Admin routes
	admin := r.Group("/api/admin")
	admin.Use(auth.AuthMiddleware(jwtSecret))
	admin.Use(auth.AdminMiddleware())
	{
		admin.GET("/users", userHandler.GetAllUsers)
		admin.GET("/users/:id", userHandler.GetUserByID)
		admin.GET("/files", fileHandler.GetAllFiles)
	}
	
	return r
}