package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhanserikAmangeldi/group-chat-service/internal/domain"
	"github.com/zhanserikAmangeldi/group-chat-service/internal/websocket"
	"gorm.io/gorm"
)

type GroupHandler struct {
	DB  *gorm.DB
	Hub *websocket.Hub
}

func NewGroupHandler(db *gorm.DB, hub *websocket.Hub) *GroupHandler {
	return &GroupHandler{DB: db, Hub: hub}
}

// CreateGroup - POST /groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name      string `json:"name"`
		CreatedBy int    `json:"created_by"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Create Group
	group := domain.Group{Name: req.Name, CreatedBy: req.CreatedBy}
	if err := h.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// 2. Add Creator as Member (Admin)
	member := domain.GroupMember{GroupID: group.ID, UserID: req.CreatedBy}
	h.DB.Create(&member)

	c.JSON(http.StatusCreated, group)
}

// JoinGroup - POST /groups/:id/join
func (h *GroupHandler) JoinGroup(c *gin.Context) {
	groupID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		UserID int `json:"user_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Verify UserID exists via User Service (gRPC/HTTP)

	member := domain.GroupMember{GroupID: groupID, UserID: req.UserID}
	if err := h.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join group"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Joined group"})
}

// ConnectGroupChat - GET /ws/groups/:id?user_id=X
func (h *GroupHandler) ConnectGroupChat(c *gin.Context) {
	groupID, _ := strconv.Atoi(c.Param("id"))
	userID, _ := strconv.Atoi(c.Query("user_id"))

	// Validate Membership
	var member domain.GroupMember
	if err := h.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this group"})
		return
	}

	// Upgrade connection
	websocket.ServeWs(h.Hub, c.Writer, c.Request, groupID, userID)
}
