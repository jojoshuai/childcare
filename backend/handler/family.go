package handler

import (
	"math/rand"
	"net/http"
	"time"

	"childcare-backend/middleware"
	"childcare-backend/model"
	"childcare-backend/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FamilyHandler handles family-related routes.
type FamilyHandler struct {
	familyStore store.FamilyStore
	userStore   store.UserStore
	inviteStore store.InviteStore
}

// NewFamilyHandler constructs a FamilyHandler.
func NewFamilyHandler(fs store.FamilyStore, us store.UserStore, is store.InviteStore) *FamilyHandler {
	return &FamilyHandler{familyStore: fs, userStore: us, inviteStore: is}
}

// GetFamily handles GET /api/family.
func (h *FamilyHandler) GetFamily(c *gin.Context) {
	familyID := middleware.GetFamilyID(c)
	if familyID == "" {
		errorResponse(c, http.StatusForbidden, "NO_FAMILY_JOINED", "请先通过邀请码加入家庭")
		return
	}

	family, err := h.familyStore.GetByID(familyID)
	if err != nil || family == nil {
		errorResponse(c, http.StatusNotFound, "FAMILY_NOT_FOUND", "家庭不存在")
		return
	}

	members, err := h.familyStore.GetMembers(familyID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取成员失败")
		return
	}

	type memberResp struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
		Role     string `json:"role"`
	}
	memberList := make([]memberResp, 0, len(members))
	for _, m := range members {
		role := ""
		if m.Role != nil {
			role = *m.Role
		}
		memberList = append(memberList, memberResp{ID: m.ID, Nickname: m.Nickname, Role: role})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      family.ID,
		"name":    family.Name,
		"members": memberList,
	})
}

const inviteChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = inviteChars[rand.Intn(len(inviteChars))]
	}
	return string(b)
}

// GenerateInvite handles POST /api/family/invite (owner only).
func (h *FamilyHandler) GenerateInvite(c *gin.Context) {
	if middleware.GetRole(c) != "owner" {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "仅家庭创建者可生成邀请码")
		return
	}

	familyID := middleware.GetFamilyID(c)
	userID := middleware.GetUserID(c)

	ic := &model.InviteCode{
		ID:        uuid.NewString(),
		FamilyID:  familyID,
		Code:      randomCode(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}
	if err := h.inviteStore.Create(ic); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建邀请码失败")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":       ic.Code,
		"expires_at": ic.ExpiresAt.Format(time.RFC3339),
	})
}

type joinRequest struct {
	Code string `json:"code" binding:"required"`
}

// JoinFamily handles POST /api/family/join.
func (h *FamilyHandler) JoinFamily(c *gin.Context) {
	var req joinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	ic, err := h.inviteStore.GetByCode(req.Code)
	if err != nil || ic == nil {
		errorResponse(c, http.StatusBadRequest, "INVITE_CODE_NOT_FOUND", "邀请码不存在")
		return
	}
	if ic.Used {
		errorResponse(c, http.StatusBadRequest, "INVITE_CODE_ALREADY_USED", "邀请码已被使用")
		return
	}
	if time.Now().After(ic.ExpiresAt) {
		errorResponse(c, http.StatusBadRequest, "INVITE_CODE_EXPIRED", "邀请码已过期")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.userStore.UpdateFamily(userID, ic.FamilyID, "member"); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "加入家庭失败")
		return
	}
	if err := h.inviteStore.MarkUsed(ic.ID); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "标记邀请码失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"family_id": ic.FamilyID})
}
