package api

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"ai-doc-system/internal/services"
)

type FriendHandler struct {
	friendService *services.FriendService
}

func NewFriendHandler(friendService *services.FriendService) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
	}
}

type SendFriendRequestRequest struct {
	ToUserID int `json:"to_user_id" binding:"required"`
}

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type AddToGroupRequest struct {
	FriendID int `json:"friend_id" binding:"required"`
	GroupID  int `json:"group_id" binding:"required"`
}

func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req SendFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.friendService.SendFriendRequest(userID.(int), req.ToUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent successfully"})
}

func (h *FriendHandler) AcceptFriendRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")
	friendIDStr := c.Param("id")
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend ID"})
		return
	}
	
	err = h.friendService.AcceptFriendRequest(userID.(int), friendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

func (h *FriendHandler) RejectFriendRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")
	friendIDStr := c.Param("id")
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend ID"})
		return
	}
	
	err = h.friendService.RejectFriendRequest(userID.(int), friendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Friend request rejected"})
}

func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	userID, _ := c.Get("user_id")
	friendIDStr := c.Param("id")
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend ID"})
		return
	}
	
	err = h.friendService.RemoveFriend(userID.(int), friendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Friend removed successfully"})
}

func (h *FriendHandler) GetFriends(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	friends, err := h.friendService.GetFriends(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get friends"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"friends": friends})
}

func (h *FriendHandler) GetPendingRequests(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	requests, err := h.friendService.GetPendingRequests(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending requests"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

func (h *FriendHandler) SearchUsers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	keyword := c.Query("keyword")
	
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keyword is required"})
		return
	}
	
	users, err := h.friendService.SearchUsers(userID.(int), keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *FriendHandler) CreateFriendGroup(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	group, err := h.friendService.CreateFriendGroup(userID.(int), req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Friend group created successfully",
		"group":   group,
	})
}

func (h *FriendHandler) GetFriendGroups(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	groups, err := h.friendService.GetFriendGroups(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get friend groups"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *FriendHandler) AddFriendToGroup(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req AddToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.friendService.AddFriendToGroup(userID.(int), req.FriendID, req.GroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Friend added to group successfully"})
}