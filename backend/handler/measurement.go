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

// MeasurementHandler handles measurement-related routes.
type MeasurementHandler struct {
	measureStore store.MeasurementStore
	childStore   store.ChildStore
}

// NewMeasurementHandler constructs a MeasurementHandler.
func NewMeasurementHandler(ms store.MeasurementStore, cs store.ChildStore) *MeasurementHandler {
	return &MeasurementHandler{measureStore: ms, childStore: cs}
}

var measureRanges = map[string][2]float64{
	"weight": {0.5, 200},
	"height": {20, 250},
}

func validateMeasureValue(mType string, value float64) bool {
	r, ok := measureRanges[mType]
	if !ok {
		return false
	}
	return value >= r[0] && value <= r[1]
}

// checkChild verifies the child exists and belongs to the current family.
func (h *MeasurementHandler) checkChild(c *gin.Context, childID string) (*model.Child, bool) {
	child, err := h.childStore.GetByID(childID)
	if err != nil || child == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "孩子不存在")
		return nil, false
	}
	return child, true
}

// List handles GET /api/children/:id/measurements.
func (h *MeasurementHandler) List(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	var mType *string
	if t := c.Query("type"); t != "" {
		mType = &t
	}

	list, err := h.measureStore.GetByChildID(childID, mType)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取记录失败")
		return
	}
	if list == nil {
		list = []*model.Measurement{}
	}
	c.JSON(http.StatusOK, list)
}

type measureRequest struct {
	Type       string  `json:"type"        binding:"required,oneof=weight height"`
	Value      float64 `json:"value"       binding:"required"`
	MeasuredAt string  `json:"measured_at" binding:"required"` // YYYY-MM-DD
	Note       *string `json:"note"`
}

// Create handles POST /api/children/:id/measurements.
func (h *MeasurementHandler) Create(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	var req measureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if !validateMeasureValue(req.Type, req.Value) {
		errorResponse(c, http.StatusBadRequest, "INVALID_VALUE", "数值超出合理范围")
		return
	}

	measuredAt, err := time.Parse("2006-01-02", req.MeasuredAt)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "measured_at 格式应为 YYYY-MM-DD")
		return
	}

	m := &model.Measurement{
		ID:         uuid.NewString(),
		ChildID:    childID,
		Type:       req.Type,
		Value:      req.Value,
		MeasuredAt: measuredAt,
		Note:       req.Note,
		CreatedBy:  middleware.GetUserID(c),
		CreatedAt:  time.Now(),
	}
	if err := h.measureStore.Create(m); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建记录失败")
		return
	}
	c.JSON(http.StatusCreated, m)
}

// Update handles PUT /api/children/:id/measurements/:mid.
func (h *MeasurementHandler) Update(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	mid := c.Param("mid")
	m, err := h.measureStore.GetByID(mid)
	if err != nil || m == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if m.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	var req measureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if !validateMeasureValue(req.Type, req.Value) {
		errorResponse(c, http.StatusBadRequest, "INVALID_VALUE", "数值超出合理范围")
		return
	}

	measuredAt, err := time.Parse("2006-01-02", req.MeasuredAt)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "measured_at 格式应为 YYYY-MM-DD")
		return
	}

	m.Type = req.Type
	m.Value = req.Value
	m.MeasuredAt = measuredAt
	m.Note = req.Note
	if err := h.measureStore.Update(m); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新失败")
		return
	}
	c.JSON(http.StatusOK, m)
}

// Delete handles DELETE /api/children/:id/measurements/:mid.
func (h *MeasurementHandler) Delete(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	mid := c.Param("mid")
	m, err := h.measureStore.GetByID(mid)
	if err != nil || m == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if m.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	if err := h.measureStore.Delete(mid); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}
