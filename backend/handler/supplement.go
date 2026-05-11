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

// SupplementHandler handles supplement-related routes.
type SupplementHandler struct {
	suppStore  store.SupplementStore
	childStore store.ChildStore
}

// NewSupplementHandler constructs a SupplementHandler.
func NewSupplementHandler(ss store.SupplementStore, cs store.ChildStore) *SupplementHandler {
	return &SupplementHandler{suppStore: ss, childStore: cs}
}

func (h *SupplementHandler) checkChild(c *gin.Context, childID string) (*model.Child, bool) {
	child, err := h.childStore.GetByID(childID)
	if err != nil || child == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "孩子不存在")
		return nil, false
	}
	return child, true
}

// List handles GET /api/children/:id/supplements.
func (h *SupplementHandler) List(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	list, err := h.suppStore.GetByChildID(childID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取补剂记录失败")
		return
	}
	if list == nil {
		list = []*model.SupplementRecord{}
	}
	c.JSON(http.StatusOK, list)
}

// GetNames handles GET /api/children/:id/supplements/names.
func (h *SupplementHandler) GetNames(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	list, err := h.suppStore.GetByChildID(childID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取补剂列表失败")
		return
	}

	// Extract unique supplement names
	seen := make(map[string]bool)
	names := []string{}
	for _, r := range list {
		if !seen[r.SupplementName] {
			seen[r.SupplementName] = true
			names = append(names, r.SupplementName)
		}
	}
	c.JSON(http.StatusOK, names)
}

type supplementRequest struct {
	SupplementName string  `json:"supplement_name" binding:"required,max=50"`
	Dose           *string `json:"dose"`
	TakenAt        string  `json:"taken_at" binding:"required"`
}

// Create handles POST /api/children/:id/supplements.
func (h *SupplementHandler) Create(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	var req supplementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	takenAt, err := time.Parse(time.RFC3339, req.TakenAt)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "taken_at 格式错误")
		return
	}

	r := &model.SupplementRecord{
		ID:             uuid.NewString(),
		ChildID:        childID,
		SupplementName: req.SupplementName,
		Dose:           req.Dose,
		TakenAt:        takenAt,
		CreatedBy:      middleware.GetUserID(c),
		CreatedAt:      time.Now(),
	}
	if err := h.suppStore.Create(r); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建记录失败")
		return
	}
	c.JSON(http.StatusCreated, r)
}

// Update handles PUT /api/children/:id/supplements/:spid.
func (h *SupplementHandler) Update(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	spid := c.Param("spid")
	r, err := h.suppStore.GetByID(spid)
	if err != nil || r == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if r.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	var req supplementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	takenAt, err := time.Parse(time.RFC3339, req.TakenAt)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "taken_at 格式错误")
		return
	}

	r.SupplementName = req.SupplementName
	r.Dose = req.Dose
	r.TakenAt = takenAt

	if err := h.suppStore.Update(r); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新失败")
		return
	}
	c.JSON(http.StatusOK, r)
}

// Delete handles DELETE /api/children/:id/supplements/:spid.
func (h *SupplementHandler) Delete(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	spid := c.Param("spid")
	r, err := h.suppStore.GetByID(spid)
	if err != nil || r == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if r.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	if err := h.suppStore.Delete(spid); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}
