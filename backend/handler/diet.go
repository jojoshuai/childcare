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

// DietHandler handles diet-related routes.
type DietHandler struct {
	dietStore  store.DietStore
	childStore store.ChildStore
}

// NewDietHandler constructs a DietHandler.
func NewDietHandler(ds store.DietStore, cs store.ChildStore) *DietHandler {
	return &DietHandler{dietStore: ds, childStore: cs}
}

func (h *DietHandler) checkChild(c *gin.Context, childID string) (*model.Child, bool) {
	child, err := h.childStore.GetByID(childID)
	if err != nil || child == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "孩子不存在")
		return nil, false
	}
	if child.FamilyID != middleware.GetFamilyID(c) {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return nil, false
	}
	return child, true
}

// List handles GET /api/children/:id/diet.
func (h *DietHandler) List(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	list, err := h.dietStore.GetByChildID(childID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取饮食记录失败")
		return
	}
	if list == nil {
		list = []*model.DietRecord{}
	}
	c.JSON(http.StatusOK, list)
}

// GetFoodTypes handles GET /api/children/:id/diet/types.
func (h *DietHandler) GetFoodTypes(c *gin.Context) {
	types := []map[string]string{
		{"value": "staple", "label": "主食"},
		{"value": "vegetable", "label": "蔬菜"},
		{"value": "fruit", "label": "水果"},
		{"value": "protein", "label": "肉蛋"},
		{"value": "dairy", "label": "奶"},
		{"value": "snack", "label": "零食"},
	}
	c.JSON(http.StatusOK, types)
}

type dietRequest struct {
	FoodName    string  `json:"food_name"    binding:"required,max=100"`
	FoodType    string  `json:"food_type"    binding:"required,oneof=staple vegetable fruit protein dairy snack"`
	AmountLevel int     `json:"amount_level" binding:"required,min=1,max=3"`
	RecordTime  string  `json:"record_time"  binding:"required"`
	Notes       *string `json:"notes"`
}

// Create handles POST /api/children/:id/diet.
func (h *DietHandler) Create(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	var req dietRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	recordTime, err := time.Parse(time.RFC3339, req.RecordTime)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "record_time 格式错误")
		return
	}

	r := &model.DietRecord{
		ID:          uuid.NewString(),
		ChildID:     childID,
		FoodName:    req.FoodName,
		FoodType:    req.FoodType,
		AmountLevel: req.AmountLevel,
		RecordTime:  recordTime,
		Notes:       req.Notes,
		CreatedBy:   middleware.GetUserID(c),
		CreatedAt:   time.Now(),
	}
	if err := h.dietStore.Create(r); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建记录失败")
		return
	}
	c.JSON(http.StatusCreated, r)
}

// Update handles PUT /api/children/:id/diet/:did.
func (h *DietHandler) Update(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	did := c.Param("did")
	r, err := h.dietStore.GetByID(did)
	if err != nil || r == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if r.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	var req dietRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	recordTime, err := time.Parse(time.RFC3339, req.RecordTime)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "record_time 格式错误")
		return
	}

	r.FoodName = req.FoodName
	r.FoodType = req.FoodType
	r.AmountLevel = req.AmountLevel
	r.RecordTime = recordTime
	r.Notes = req.Notes

	if err := h.dietStore.Update(r); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新失败")
		return
	}
	c.JSON(http.StatusOK, r)
}

// Delete handles DELETE /api/children/:id/diet/:did.
func (h *DietHandler) Delete(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	did := c.Param("did")
	r, err := h.dietStore.GetByID(did)
	if err != nil || r == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if r.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	if err := h.dietStore.Delete(did); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}
