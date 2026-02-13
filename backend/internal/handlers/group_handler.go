package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"onechat/internal/services"
	"onechat/internal/websocket"
)

type GroupHandler struct {
	groupService *services.GroupService
	hub          *websocket.Hub
}

func NewGroupHandler(groupService *services.GroupService, hub *websocket.Hub) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		hub:          hub,
	}
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	MemberIDs   []uint `json:"member_ids"`
}

type AddMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.groupService.CreateGroup(req.Name, req.Description, req.Icon, userID, req.MemberIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"group": group})
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	group, err := h.groupService.GetGroup(uint(groupID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove protected fields
	delete(updates, "id")
	delete(updates, "created_by_id")
	delete(updates, "created_at")

	group, err := h.groupService.UpdateGroup(uint(groupID), userID, updates)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Broadcast update to group members
	updateNotif, _ := json.Marshal(map[string]interface{}{
		"type":  "group_updated",
		"group": group,
	})
	h.hub.BroadcastToChat(uint(groupID), updateNotif, 0)

	c.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	if err := h.groupService.DeleteGroup(uint(groupID), userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *GroupHandler) AddMember(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupService.AddMember(uint(groupID), userID, req.UserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Broadcast member addition
	memberNotif, _ := json.Marshal(map[string]interface{}{
		"type":     "member_added",
		"group_id": groupID,
		"user_id":  req.UserID,
	})
	h.hub.BroadcastToChat(uint(groupID), memberNotif, 0)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	memberID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.groupService.RemoveMember(uint(groupID), userID, uint(memberID)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Broadcast member removal
	removeNotif, _ := json.Marshal(map[string]interface{}{
		"type":     "member_removed",
		"group_id": groupID,
		"user_id":  memberID,
	})
	h.hub.BroadcastToChat(uint(groupID), removeNotif, 0)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	memberID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupService.UpdateMemberRole(uint(groupID), userID, uint(memberID), req.Role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Broadcast role update
	roleNotif, _ := json.Marshal(map[string]interface{}{
		"type":     "role_updated",
		"group_id": groupID,
		"user_id":  memberID,
		"role":     req.Role,
	})
	h.hub.BroadcastToChat(uint(groupID), roleNotif, 0)

	c.JSON(http.StatusOK, gin.H{"success": true})
}
