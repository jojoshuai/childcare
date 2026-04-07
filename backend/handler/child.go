package handler

import (
	"net/http"
	"time"

	"childcare-backend/middleware"
	"childcare-backend/model"
	"childcare-backend/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChildHandler handles child-related routes.
type ChildHandler struct {
	childStore store.ChildStore
}

// NewChildHandler constructs a ChildHandler.
func NewChildHandler(cs store.ChildStore) *ChildHandler {
	return &ChildHandler{childStore: cs}
}

// List handles GET /api/children.
func (h *ChildHandler) List(c *gin.Context) {
	familyID := middleware.GetFamilyID(c)
	children, err := h.childStore.GetByFamilyID(familyID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取孩子列表失败")
		return
	}
	if children == nil {
		children = []*model.Child{}
	}
	c.JSON(http.StatusOK, children)
}

type childRequest struct {
	Name      string `json:"name"       binding:"required,min=1,max=50"`
	Gender    string `json:"gender"     binding:"required,oneof=male female"`
	BirthDate string `json:"birth_date" binding:"required"` // YYYY-MM-DD
}

// Create handles POST /api/children.
func (h *ChildHandler) Create(c *gin.Context) {
	var req childRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "birth_date 格式应为 YYYY-MM-DD")
		return
	}

	child := &model.Child{
		ID:        uuid.NewString(),
		FamilyID:  middleware.GetFamilyID(c),
		Name:      req.Name,
		Gender:    req.Gender,
		BirthDate: birthDate,
		CreatedAt: time.Now(),
	}
	if err := h.childStore.Create(child); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建孩子失败")
		return
	}
	c.JSON(http.StatusCreated, child)
}

// Update handles PUT /api/children/:id.
func (h *ChildHandler) Update(c *gin.Context) {
	id := c.Param("id")
	child, err := h.childStore.GetByID(id)
	if err != nil || child == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "孩子不存在")
		return
	}
	if child.FamilyID != middleware.GetFamilyID(c) {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	var req childRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "birth_date 格式应为 YYYY-MM-DD")
		return
	}

	child.Name = req.Name
	child.Gender = req.Gender
	child.BirthDate = birthDate
	if err := h.childStore.Update(child); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新失败")
		return
	}
	c.JSON(http.StatusOK, child)
}

// Delete handles DELETE /api/children/:id (owner only).
func (h *ChildHandler) Delete(c *gin.Context) {
	if middleware.GetRole(c) != "owner" {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "仅家庭创建者可删除孩子")
		return
	}
	id := c.Param("id")
	child, err := h.childStore.GetByID(id)
	if err != nil || child == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "孩子不存在")
		return
	}
	if child.FamilyID != middleware.GetFamilyID(c) {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}
	if err := h.childStore.Delete(id); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}
