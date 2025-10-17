package api

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"ai-doc-system/internal/services"
)

type FileShareHandler struct {
	fileShareService *services.FileShareService
	fileService      *services.FileService
}

func NewFileShareHandler(fileShareService *services.FileShareService, fileService *services.FileService) *FileShareHandler {
	return &FileShareHandler{
		fileShareService: fileShareService,
		fileService:      fileService,
	}
}

type ShareToFriendRequest struct {
	FileID   int `json:"file_id" binding:"required"`
	FriendID int `json:"friend_id" binding:"required"`
}

type CreatePublicShareRequest struct {
	FileID    int    `json:"file_id" binding:"required"`
	ExpiresIn *int   `json:"expires_in"` // Expiration time (hours)
}

func (h *FileShareHandler) ShareToFriend(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req ShareToFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.fileShareService.ShareFileToFriend(req.FileID, userID.(int), req.FriendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "File shared successfully"})
}

func (h *FileShareHandler) CreatePublicShare(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req CreatePublicShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &expiry
	}
	
	share, err := h.fileShareService.CreatePublicShare(req.FileID, userID.(int), expiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Public share created successfully",
		"share":   share,
	})
}

func (h *FileShareHandler) GetSharedWithMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	files, err := h.fileShareService.GetSharedWithMeFiles(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get shared files"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileShareHandler) GetMyShares(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	files, err := h.fileShareService.GetMySharedFiles(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get shared files"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileShareHandler) DownloadSharedFile(c *gin.Context) {
	shareToken := c.Param("token")
	
	file, err := h.fileShareService.GetFileByShareToken(shareToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share link not found or expired"})
		return
	}
	
	c.Header("Content-Disposition", "attachment; filename="+file.Filename)
	c.Header("Content-Type", file.MimeType)
	c.File(file.FilePath)
}

func (h *FileShareHandler) DownloadFriendSharedFile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	
	// Check access permission
	hasAccess, err := h.fileShareService.CheckFileAccess(fileID, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access"})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	
	file, err := h.fileService.GetFileByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	c.Header("Content-Disposition", "attachment; filename="+file.Filename)
	c.Header("Content-Type", file.MimeType)
	c.File(file.FilePath)
}

func (h *FileShareHandler) RemoveShare(c *gin.Context) {
	userID, _ := c.Get("user_id")
	shareIDStr := c.Param("id")
	shareID, err := strconv.Atoi(shareIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}
	
	err = h.fileShareService.RemoveShare(shareID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Share removed successfully"})
}