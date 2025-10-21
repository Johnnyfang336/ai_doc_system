package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"ai-doc-system/internal/auth"
	"ai-doc-system/internal/services"
)

type OnlyOfficeHandler struct {
	onlyOfficeURL string
	callbackURL   string
	jwtSecret     string
	uploadPath    string
	fileService   *services.FileService
}

func NewOnlyOfficeHandler(jwtSecret string, fileService *services.FileService) *OnlyOfficeHandler {
	return &OnlyOfficeHandler{
		onlyOfficeURL: "http://onlyoffice:80", // Internal Docker network URL
		callbackURL:   "http://backend:8080/api/onlyoffice/callback",
		jwtSecret:     jwtSecret,
		uploadPath:    "./storage/files",
		fileService:   fileService,
	}
}

// OnlyOffice document configuration structure
type DocumentConfig struct {
	Document struct {
		FileType    string            `json:"fileType"`
		Key         string            `json:"key"`
		Title       string            `json:"title"`
		URL         string            `json:"url"`
		Permissions map[string]bool   `json:"permissions"`
		Info        map[string]string `json:"info,omitempty"`
	} `json:"document"`
	DocumentType string `json:"documentType"`
	EditorConfig struct {
		CallbackURL string `json:"callbackUrl"`
		Lang        string `json:"lang"`
		Mode        string `json:"mode"`
		User        struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
	} `json:"editorConfig"`
	Height string `json:"height"`
	Type   string `json:"type"`
	Width  string `json:"width"`
}

// GetOnlyOfficeConfig generates OnlyOffice configuration for a file
func (h *OnlyOfficeHandler) GetOnlyOfficeConfig(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Parse token from query or Authorization header
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token required"})
		return
	}

	claims, err := auth.ValidateToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Fetch real file info instead of mock; enforce ownership
	file, err := h.fileService.GetFileByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	if file.UserID != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	filename := file.OriginalName
	if filename == "" {
		filename = file.Filename
	}
	fileType := getFileTypeFromExtension(filename)
	documentType := getDocumentType(fileType)

	config := DocumentConfig{
		DocumentType: documentType,
		Height:       "100%",
		Type:         "desktop",
		Width:        "100%",
	}

	// Document configuration
	config.Document.FileType = fileType
	config.Document.Key = fmt.Sprintf("file_%d_%d", fileID, claims.UserID)
	config.Document.Title = filename
	config.Document.URL = fmt.Sprintf("http://backend:8080/api/files/%d/download?token=%s", fileID, token)
	config.Document.Permissions = map[string]bool{
		"comment":              true,
		"copy":                 true,
		"download":             true,
		"edit":                 true,
		"fillForms":            true,
		"modifyFilter":         true,
		"modifyContentControl": true,
		"review":               true,
		"reviewGroups":         true,
		"chat":                 true,
		"commentGroups":        true,
		"userInfoGroups":       true,
		"protect":              true,
	}

	// Editor configuration
	config.EditorConfig.CallbackURL = h.callbackURL
	config.EditorConfig.Lang = "en"
	config.EditorConfig.Mode = "edit"
	config.EditorConfig.User.ID = fmt.Sprintf("%d", claims.UserID)
	config.EditorConfig.User.Name = claims.Username

	c.JSON(http.StatusOK, config)
}

// HandleCallback handles OnlyOffice document save callbacks
func (h *OnlyOfficeHandler) HandleCallback(c *gin.Context) {
	var callbackData map[string]interface{}
	if err := c.ShouldBindJSON(&callbackData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback data"})
		return
	}

	// Log callback for debugging
	fmt.Printf("OnlyOffice callback received: %+v\n", callbackData)

	// Handle different callback statuses
	status, ok := callbackData["status"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	switch int(status) {
	case 1: // Document is being edited
		c.JSON(http.StatusOK, gin.H{"error": 0})
	case 2: // Document is ready for saving
		// In production, save the document here
		downloadURL, exists := callbackData["url"].(string)
		if exists {
			fmt.Printf("Document ready for saving from: %s\n", downloadURL)
			
			// Get file ID from the request
			fileIDStr := c.Param("id")
			if fileIDStr == "" {
				fmt.Printf("Error: No file ID provided in callback\n")
				c.JSON(http.StatusOK, gin.H{"error": 1})
				return
			}
			
			// Download the document from OnlyOffice
			resp, err := http.Get(downloadURL)
			if err != nil {
				fmt.Printf("Error downloading document: %v\n", err)
				c.JSON(http.StatusOK, gin.H{"error": 1})
				return
			}
			defer resp.Body.Close()
			
			// Create the file path
			filePath := filepath.Join(h.uploadPath, fileIDStr)
			
			// Create the file
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				c.JSON(http.StatusOK, gin.H{"error": 1})
				return
			}
			defer file.Close()
			
			// Copy the content
			_, err = io.Copy(file, resp.Body)
			if err != nil {
				fmt.Printf("Error saving document: %v\n", err)
				c.JSON(http.StatusOK, gin.H{"error": 1})
				return
			}
			
			fmt.Printf("Document saved successfully to: %s\n", filePath)
		}
		c.JSON(http.StatusOK, gin.H{"error": 0})
	case 3: // Document saving error
		c.JSON(http.StatusOK, gin.H{"error": 0})
	case 4: // Document closed with no changes
		c.JSON(http.StatusOK, gin.H{"error": 0})
	case 6: // Document is being edited, but the current document state is saved
		c.JSON(http.StatusOK, gin.H{"error": 0})
	case 7: // Error has occurred while force saving the document
		c.JSON(http.StatusOK, gin.H{"error": 0})
	default:
		c.JSON(http.StatusOK, gin.H{"error": 0})
	}
}

// Helper functions
func getFileTypeFromExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return "txt"
	}
	ext := strings.ToLower(parts[len(parts)-1])
	
	// Map extensions to OnlyOffice supported types
	switch ext {
	case "doc", "docx":
		return "docx"
	case "xls", "xlsx":
		return "xlsx"
	case "ppt", "pptx":
		return "pptx"
	case "pdf":
		return "pdf"
	case "txt":
		return "txt"
	default:
		return "txt"
	}
}

func getDocumentType(fileType string) string {
	switch fileType {
	case "docx", "doc", "txt":
		return "word"
	case "xlsx", "xls":
		return "cell"
	case "pptx", "ppt":
		return "slide"
	default:
		return "word"
	}
}