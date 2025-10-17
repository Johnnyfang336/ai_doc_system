package api

import (
	"net/http"
	"path/filepath"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"ai-doc-system/internal/services"
)

type FileHandler struct {
	fileService *services.FileService
}

func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

type RenameFileRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	
	uploadedFile, err := h.fileService.UploadFile(userID.(int), file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"file":    uploadedFile,
	})
}

func (h *FileHandler) GetUserFiles(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	files, err := h.fileService.GetUserFiles(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get files"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileHandler) GetFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	
	file, err := h.fileService.GetFileByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"file": file})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	
	file, err := h.fileService.GetFileByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	// Check if file exists
	if !filepath.IsAbs(file.FilePath) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid file path"})
		return
	}
	
	c.Header("Content-Disposition", "attachment; filename="+file.Filename)
	c.Header("Content-Type", file.MimeType)
	c.File(file.FilePath)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	
	err = h.fileService.DeleteFile(fileID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func (h *FileHandler) RenameFile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	
	var req RenameFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err = h.fileService.RenameFile(fileID, userID.(int), req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "File renamed successfully"})
}

func (h *FileHandler) GetAllFiles(c *gin.Context) {
	files, err := h.fileService.GetAllFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get files"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileHandler) GetStorageUsage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	usage, err := h.fileService.GetUserStorageUsage(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get storage usage"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"used":  usage,
		"limit": 100 * 1024 * 1024, // 100MB
		"percentage": float64(usage) / (100 * 1024 * 1024) * 100,
	})
}