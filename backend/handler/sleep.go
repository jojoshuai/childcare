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

// SleepHandler handles sleep-related routes.
type SleepHandler struct {
	sleepStore store.SleepStore
	childStore store.ChildStore
}

// NewSleepHandler constructs a SleepHandler.
func NewSleepHandler(ss store.SleepStore, cs store.ChildStore) *SleepHandler {
	return &SleepHandler{sleepStore: ss, childStore: cs}
}

func (h *SleepHandler) checkChild(c *gin.Context, childID string) (*model.Child, bool) {
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

// List handles GET /api/children/:id/sleep.
func (h *SleepHandler) List(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	list, err := h.sleepStore.GetByChildID(childID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取睡眠记录失败")
		return
	}
	if list == nil {
		list = []*model.SleepRecord{}
	}
	c.JSON(http.StatusOK, list)
}

type sleepRequest struct {
	StartTime string  `json:"start_time" binding:"required"`
	EndTime   *string `json:"end_time"`
	WokeUp    bool    `json:"woke_up"`
	WakeCount int     `json:"wake_count"`
}

// Create handles POST /api/children/:id/sleep.
func (h *SleepHandler) Create(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	var req sleepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "start_time 格式错误")
		return
	}

	var endTime *time.Time
	if req.EndTime != nil {
		t, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "end_time 格式错误")
			return
		}
		endTime = &t
	}

	r := &model.SleepRecord{
		ID:        uuid.NewString(),
		ChildID:   childID,
		StartTime: startTime,
		EndTime:   endTime,
		WokeUp:    req.WokeUp,
		WakeCount: req.WakeCount,
		CreatedBy: middleware.GetUserID(c),
		CreatedAt: time.Now(),
	}
	if err := h.sleepStore.Create(r); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建记录失败")
		return
	}
	c.JSON(http.StatusCreated, r)
}

// Update handles PUT /api/children/:id/sleep/:sid.
func (h *SleepHandler) Update(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	sid := c.Param("sid")
	r, err := h.sleepStore.GetByID(sid)
	if err != nil || r == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if r.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	var req sleepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "start_time 格式错误")
		return
	}

	r.StartTime = startTime
	if req.EndTime != nil {
		t, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "INVALID_DATE", "end_time 格式错误")
			return
		}
		r.EndTime = &t
	} else {
		r.EndTime = nil
	}
	r.WokeUp = req.WokeUp
	r.WakeCount = req.WakeCount

	if err := h.sleepStore.Update(r); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新失败")
		return
	}
	c.JSON(http.StatusOK, r)
}

// Delete handles DELETE /api/children/:id/sleep/:sid.
func (h *SleepHandler) Delete(c *gin.Context) {
	childID := c.Param("id")
	if _, ok := h.checkChild(c, childID); !ok {
		return
	}

	sid := c.Param("sid")
	r, err := h.sleepStore.GetByID(sid)
	if err != nil || r == nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "记录不存在")
		return
	}
	if r.ChildID != childID {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "无权操作")
		return
	}

	if err := h.sleepStore.Delete(sid); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}
