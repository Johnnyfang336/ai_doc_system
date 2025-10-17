package api

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"ai-doc-system/internal/services"
)

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

type SendMessageRequest struct {
	ToUserID int    `json:"to_user_id" binding:"required"`
	Content  string `json:"content" binding:"required,min=1,max=1000"`
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	message, err := h.messageService.SendMessage(userID.(int), req.ToUserID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Message sent successfully",
		"data":    message,
	})
}

func (h *MessageHandler) GetChatHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	friendIDStr := c.Param("friend_id")
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend ID"})
		return
	}
	
	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	
	messages, err := h.messageService.GetChatHistory(userID.(int), friendID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat history"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) GetChatList(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	chatList, err := h.messageService.GetChatList(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat list"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"chats": chatList})
}

func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("user_id")
	fromUserIDStr := c.Param("from_user_id")
	fromUserID, err := strconv.Atoi(fromUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	err = h.messageService.MarkMessagesAsRead(userID.(int), fromUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read"})
}

func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	count, err := h.messageService.GetUnreadCount(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	messageIDStr := c.Param("id")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}
	
	err = h.messageService.DeleteMessage(messageID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}