# Backend Handlers & Main 实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现后端剩余的 HTTP Handler 和 main.go 入口，使后端 API 完整可运行。

**Architecture:** 在已有的 store/middleware/model 基础上，依次实现 family、child、measurement、who-standards 四组 handler，最后用 main.go 将路由全部注册并启动服务。每个 handler 依赖对应的 store 接口，通过依赖注入传入。

**Tech Stack:** Go + Gin, JWT, golang-migrate, MySQL (go-sql-driver), github.com/google/uuid

---

## Chunk 1: main.go + family handler

### Task 1: 创建 `backend/main.go`

**Files:**
- Create: `backend/main.go`

- [ ] **Step 1: 写 main.go**

```go
package main

import (
	"log"

	"childcare-backend/config"
	"childcare-backend/db"
	"childcare-backend/handler"
	"childcare-backend/middleware"
	"childcare-backend/store"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database, cfg); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// stores
	userStore    := store.NewUserStore(database)
	familyStore  := store.NewFamilyStore(database)
	childStore   := store.NewChildStore(database)
	measureStore := store.NewMeasurementStore(database)
	inviteStore  := store.NewInviteStore(database)

	// handlers
	authH    := handler.NewAuthHandler(userStore, familyStore, cfg)
	familyH  := handler.NewFamilyHandler(familyStore, userStore, inviteStore)
	childH   := handler.NewChildHandler(childStore)
	measureH := handler.NewMeasurementHandler(measureStore, childStore)

	r := gin.Default()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login",    authH.Login)
			auth.POST("/wx-login", authH.WxLogin)
			auth.POST("/refresh",  authH.Refresh)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/family",         familyH.GetFamily)
			protected.POST("/family/invite",  familyH.GenerateInvite)
			protected.POST("/family/join",    familyH.JoinFamily)

			withFamily := protected.Group("")
			withFamily.Use(middleware.RequireFamily())
			{
				withFamily.GET("/children",                              childH.List)
				withFamily.POST("/children",                             childH.Create)
				withFamily.PUT("/children/:id",                          childH.Update)
				withFamily.DELETE("/children/:id",                       childH.Delete)

				withFamily.GET("/children/:id/measurements",             measureH.List)
				withFamily.POST("/children/:id/measurements",            measureH.Create)
				withFamily.PUT("/children/:id/measurements/:mid",        measureH.Update)
				withFamily.DELETE("/children/:id/measurements/:mid",     measureH.Delete)

				withFamily.GET("/who-standards", handler.GetWHOStandards)
			}
		}
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("run: %v", err)
	}
}
```

- [ ] **Step 2: 确认 config.Config 有 Port 字段（若无则补加）**

查看 `backend/config/config.go`，确认有 `Port string`。如没有则补加并读取环境变量 `PORT`，默认 `"8080"`。

- [ ] **Step 3: 确认 db 包有 Migrate 函数**

查看 `backend/db/db.go`，确认有 `Migrate(db *sql.DB, cfg *config.Config) error`。

---

### Task 2: 创建 `backend/handler/family.go`

**Files:**
- Create: `backend/handler/family.go`

接口：
- `GET /api/family` — 获取家庭信息及成员列表（需已加入家庭）
- `POST /api/family/invite` — 生成邀请码（仅 owner，6位，24小时有效）
- `POST /api/family/join` — 用邀请码加入家庭

- [ ] **Step 1: 写 family.go**

```go
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

type FamilyHandler struct {
	familyStore store.FamilyStore
	userStore   store.UserStore
	inviteStore store.InviteStore
}

func NewFamilyHandler(fs store.FamilyStore, us store.UserStore, is store.InviteStore) *FamilyHandler {
	return &FamilyHandler{familyStore: fs, userStore: us, inviteStore: is}
}

// GET /api/family
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

// POST /api/family/invite
func (h *FamilyHandler) GenerateInvite(c *gin.Context) {
	role := middleware.GetRole(c)
	if role != "owner" {
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "仅家庭创建者可生成邀请码")
		return
	}

	familyID := middleware.GetFamilyID(c)
	userID   := middleware.GetUserID(c)

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

// POST /api/family/join
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
```

---

## Chunk 2: child + measurement handler

### Task 3: 创建 `backend/handler/child.go`

**Files:**
- Create: `backend/handler/child.go`

- [ ] **Step 1: 写 child.go**

```go
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

type ChildHandler struct {
	childStore store.ChildStore
}

func NewChildHandler(cs store.ChildStore) *ChildHandler {
	return &ChildHandler{childStore: cs}
}

// GET /api/children
func (h *ChildHandler) List(c *gin.Context) {
	familyID := middleware.GetFamilyID(c)
	children, err := h.childStore.GetByFamilyID(familyID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取孩子列表失败")
		return
	}
	c.JSON(http.StatusOK, children)
}

type childRequest struct {
	Name      string `json:"name"       binding:"required,min=1,max=50"`
	Gender    string `json:"gender"     binding:"required,oneof=male female"`
	BirthDate string `json:"birth_date" binding:"required"` // YYYY-MM-DD
}

// POST /api/children
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

// PUT /api/children/:id
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

	child.Name      = req.Name
	child.Gender    = req.Gender
	child.BirthDate = birthDate
	if err := h.childStore.Update(child); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新失败")
		return
	}
	c.JSON(http.StatusOK, child)
}

// DELETE /api/children/:id（仅 owner）
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
```

---

### Task 4: 创建 `backend/handler/measurement.go`

**Files:**
- Create: `backend/handler/measurement.go`

测量值范围校验：weight 0.5–200 kg；height 20–250 cm；head_circumference 20–80 cm。

- [ ] **Step 1: 写 measurement.go**

```go
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

type MeasurementHandler struct {
	measureStore store.MeasurementStore
	childStore   store.ChildStore
}

func NewMeasurementHandler(ms store.MeasurementStore, cs store.ChildStore) *MeasurementHandler {
	return &MeasurementHandler{measureStore: ms, childStore: cs}
}

var measureRanges = map[string][2]float64{
	"weight":            {0.5, 200},
	"height":            {20, 250},
	"head_circumference": {20, 80},
}

func validateMeasureValue(mType string, value float64) bool {
	r, ok := measureRanges[mType]
	if !ok {
		return false
	}
	return value >= r[0] && value <= r[1]
}

// 验证 child 属于当前家庭
func (h *MeasurementHandler) checkChild(c *gin.Context, childID string) (*model.Child, bool) {
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

// GET /api/children/:id/measurements
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
	c.JSON(http.StatusOK, list)
}

type measureRequest struct {
	Type       string  `json:"type"        binding:"required,oneof=weight height head_circumference"`
	Value      float64 `json:"value"       binding:"required"`
	MeasuredAt string  `json:"measured_at" binding:"required"` // YYYY-MM-DD
	Note       *string `json:"note"`
}

// POST /api/children/:id/measurements
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

// PUT /api/children/:id/measurements/:mid
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

	m.Type       = req.Type
	m.Value      = req.Value
	m.MeasuredAt = measuredAt
	m.Note       = req.Note
	if err := h.measureStore.Update(m); err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新失败")
		return
	}
	c.JSON(http.StatusOK, m)
}

// DELETE /api/children/:id/measurements/:mid
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
```

---

## Chunk 3: WHO 数据 + 验证编译

### Task 5: 创建 `backend/handler/who.go`

**Files:**
- Create: `backend/handler/who.go`

WHO 数据内嵌在代码中，不存数据库。P3/P50/P97，0-60月龄，按性别 × 类型划分。

- [ ] **Step 1: 写 who.go（含内嵌 WHO 数据）**

见代码实现，数据来源 WHO Child Growth Standards。

---

### Task 6: 验证编译

- [ ] **Step 1: 进入 backend 目录，执行 `go build ./...`**

```bash
cd backend && go build ./...
```

预期：无错误输出。

- [ ] **Step 2: 执行 `go vet ./...`**

```bash
go vet ./...
```

预期：无警告。

- [ ] **Step 3: 执行测试**

```bash
go test ./...
```

预期：PASS（已有测试通过，新代码无额外测试要求）。

- [ ] **Step 4: Commit**

```bash
git add backend/main.go backend/handler/family.go backend/handler/child.go \
        backend/handler/measurement.go backend/handler/who.go
git commit -m "feat: add family/child/measurement/WHO handlers and main.go"
```
