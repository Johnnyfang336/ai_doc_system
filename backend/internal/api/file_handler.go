package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	
	"github.com/gin-gonic/gin"
	"ai-doc-system/internal/models"
	"ai-doc-system/internal/services"
	"ai-doc-system/internal/auth"
)

type FileHandler struct {
	fileService *services.FileService
	jwtSecret   string
}

func NewFileHandler(fileService *services.FileService, jwtSecret string) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		jwtSecret:   jwtSecret,
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
	fmt.Printf("DownloadFile called for path: %s\n", c.Request.URL.Path)
	fmt.Printf("Authorization header: %s\n", c.GetHeader("Authorization"))
	
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		fmt.Printf("File ID parameter: %s\n", fileIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	fmt.Printf("File ID parameter: %d\n", fileID)

	// Check authentication - either from header or query parameter
	var authenticated bool
	var userID interface{}
	
	// First try to get user_id from context (set by middleware)
	if uid, exists := c.Get("user_id"); exists {
		authenticated = true
		userID = uid
		fmt.Printf("User ID from context: %v\n", userID)
	} else {
		// If not authenticated via middleware, try token from query parameter
		tokenFromQuery := c.Query("token")
		if tokenFromQuery != "" {
			fmt.Printf("Trying token from query: %s\n", tokenFromQuery)
			// Validate token from query parameter
			claims, err := auth.ValidateToken(tokenFromQuery, h.jwtSecret)
			if err == nil {
				authenticated = true
				userID = claims.UserID
				fmt.Printf("Token validated, user ID: %v\n", userID)
				// Set context for consistency
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("role", claims.Role)
			} else {
				fmt.Printf("Token validation failed: %v\n", err)
			}
		}
	}
	
	if !authenticated {
		fmt.Printf("User not authenticated\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	fmt.Printf("Getting file by ID: %d\n", fileID)
	file, err := h.fileService.GetFileByID(fileID)
	if err != nil {
		fmt.Printf("File not found error: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	fmt.Printf("File found: %+v\n", file)
	// Check if user owns the file
	if file.UserID != userID.(int) {
		fmt.Printf("Access denied - file.UserID: %d, userID: %d\n", file.UserID, userID.(int))
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	
	// Check if file exists
	if file.FilePath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid file path"})
		return
	}
	
	// Convert relative path to absolute path if needed
	var fullPath string
	if filepath.IsAbs(file.FilePath) {
		fullPath = file.FilePath
	} else {
		// Assume relative path is from the working directory
		fullPath = filepath.Join(".", file.FilePath)
	}
	
	c.Header("Content-Disposition", "attachment; filename="+file.Filename)
	c.Header("Content-Type", file.MimeType)
	c.Header("X-Filename", file.Filename)
	c.File(fullPath)
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

// EditFile provides online file editing functionality
func (h *FileHandler) EditFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	
	// Check authentication - either from header or query parameter
	var authenticated bool
	
	// First try to get user_id from context (set by middleware)
	if _, exists := c.Get("user_id"); exists {
		authenticated = true
	} else {
		// If not authenticated via middleware, try token from query parameter
		tokenFromQuery := c.Query("token")
		if tokenFromQuery != "" {
			// Validate token from query parameter
			claims, err := auth.ValidateToken(tokenFromQuery, h.jwtSecret)
			if err == nil {
				authenticated = true
				// Set context for consistency
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("role", claims.Role)
			}
		}
	}
	
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	file, err := h.fileService.GetFileByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	// Check if file type supports online editing
	if !isSupportedFileType(file.MimeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not supported for online editing"})
		return
	}
	
	// Return appropriate editor based on file type
	editorType := getEditorType(file.MimeType)
	
	// Build editor HTML page
	editorHTML := generateEditorHTML(file, editorType)
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, editorHTML)
}

// PreviewFile provides file preview functionality
func (h *FileHandler) PreviewFile(c *gin.Context) {
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
	
	// For PDF files, return file content directly
	if file.MimeType == "application/pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "inline; filename="+file.Filename)
		c.File(file.FilePath)
		return
	}
	
	// For other file types, return preview page
	previewHTML := generatePreviewHTML(file)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, previewHTML)
}

// Check if file type supports online editing
func isSupportedFileType(mimeType string) bool {
	supportedTypes := []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document", // .docx
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",       // .xlsx
		"application/vnd.openxmlformats-officedocument.presentationml.presentation", // .pptx
		"application/msword",        // .doc
		"application/vnd.ms-excel",  // .xls
		"application/vnd.ms-powerpoint", // .ppt
		"application/pdf",           // .pdf
		"text/plain",               // .txt
		"application/vnd.oasis.opendocument.text",         // .odt
		"application/vnd.oasis.opendocument.spreadsheet",  // .ods
		"application/vnd.oasis.opendocument.presentation", // .odp
	}
	
	for _, supportedType := range supportedTypes {
		if mimeType == supportedType {
			return true
		}
	}
	return false
}

// Get editor type based on file mime type
func getEditorType(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "word") || strings.Contains(mimeType, "document"):
		return "document"
	case strings.Contains(mimeType, "excel") || strings.Contains(mimeType, "spreadsheet"):
		return "spreadsheet"
	case strings.Contains(mimeType, "powerpoint") || strings.Contains(mimeType, "presentation"):
		return "presentation"
	case mimeType == "application/pdf":
		return "pdf"
	case mimeType == "text/plain":
		return "text"
	default:
		return "document"
	}
}

// Generate editor HTML page
func generateEditorHTML(file *models.File, editorType string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Edit - %s</title>
    <script type="text/javascript" src="http://localhost:8081/web-apps/apps/api/documents/api.js"></script>
    <style>
        html, body { height: 100%%; margin: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif; background-color: #ffffff; }
        #onlyoffice-editor { width: 100%%; height: 100vh; }
    </style>
</head>
<body>
    <div id="onlyoffice-editor"></div>
    <script>
        function getTokenFromURL() {
            const urlParams = new URLSearchParams(window.location.search);
            return urlParams.get('token');
        }

        async function initEditor() {
            const token = getTokenFromURL();
            if (!token) {
                document.body.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100vh;color:#d32f2f;">Authentication token not found</div>';
                return;
            }
            const fileId = window.location.pathname.split('/')[3];

            try {
                const resp = await fetch('/api/files/' + fileId + '/onlyoffice/config', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                if (!resp.ok) throw new Error('Failed to get OnlyOffice configuration');
                const config = await resp.json();

                if (typeof DocsAPI !== 'undefined') {
                    new DocsAPI.DocEditor('onlyoffice-editor', config);
                } else {
                    document.getElementById('onlyoffice-editor').innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%%;color:#666;">OnlyOffice API not available</div>';
                }
            } catch (e) {
                console.error(e);
                document.getElementById('onlyoffice-editor').innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%%;color:#d32f2f;">' + e.message + '</div>';
            }
        }

        document.addEventListener('DOMContentLoaded', initEditor);
    </script>
</body>
</html>
`, file.Filename)
}

// Generate preview HTML page
func generatePreviewHTML(file *models.File) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Preview - %s</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
            background-color: #f5f5f5;
        }
        .preview-container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 20px;
        }
        .header {
            border-bottom: 1px solid #e0e0e0;
            padding-bottom: 15px;
            margin-bottom: 20px;
        }
        .content {
            text-align: center;
            padding: 40px 20px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="preview-container">
        <div class="header">
            <h2>%s</h2>
            <p>File Type: %s</p>
        </div>
        <div class="content">
            <h3>File Preview</h3>
            <p>Preview functionality for this file type is under development.</p>
            <p>You can download the file to view it.</p>
        </div>
    </div>
</body>
</html>
`, file.Filename, file.Filename, file.MimeType)
}