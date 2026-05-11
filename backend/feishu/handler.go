package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"childcare-backend/model"
	"childcare-backend/store"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkdispatcher "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/google/uuid"
)

// Bot represents the Feishu bot integration.
type Bot struct {
	larkClient   *lark.Client
	appID        string
	appSecret    string
	llm          *Client
	childStore   store.ChildStore
	sleepStore   store.SleepStore
	dietStore    store.DietStore
	suppStore    store.SupplementStore
	measureStore store.MeasurementStore
}

// NewBot creates a new Feishu bot instance.
func NewBot(appID, appSecret, deepseekAPIKey string,
	childStore store.ChildStore,
	sleepStore store.SleepStore, dietStore store.DietStore,
	suppStore store.SupplementStore, measureStore store.MeasurementStore,
) *Bot {
	return &Bot{
		larkClient:   lark.NewClient(appID, appSecret),
		appID:        appID,
		appSecret:    appSecret,
		llm:          NewClient(deepseekAPIKey),
		childStore:   childStore,
		sleepStore:   sleepStore,
		dietStore:    dietStore,
		suppStore:    suppStore,
		measureStore: measureStore,
	}
}

// StartWS starts the Feishu WebSocket long connection.
func (b *Bot) StartWS() error {
	eventDispatcher := larkdispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			go b.handleMessage(ctx, event)
			return nil
		}).
		OnP2ChatAccessEventBotP2pChatEnteredV1(func(ctx context.Context, event *larkim.P2ChatAccessEventBotP2pChatEnteredV1) error {
			log.Printf("[DEBUG][feishu] bot_p2p_chat_entered_v1 received, ignoring")
			return nil
		}).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			// Message read event, no action needed
			return nil
		})

	log.Printf("[DEBUG][feishu] registered event handlers, starting WS connection...")

	wsClient := larkws.NewClient(b.appID, b.appSecret,
		larkws.WithEventHandler(eventDispatcher),
		larkws.WithLogLevel(larkcore.LogLevelDebug),
	)

	err := wsClient.Start(context.Background())
	if err != nil {
		return fmt.Errorf("feishu ws start: %w", err)
	}
	return nil
}

// ParsedRecord represents a single parsed record from LLM.
type ParsedRecord struct {
	Category       string   `json:"category"`
	MealGroupID    *string  `json:"meal_group_id"`
	MealType       *string  `json:"meal_type"`
	StartTime      *string  `json:"start_time"`
	EndTime        *string  `json:"end_time"`
	WokeUp         *bool    `json:"woke_up"`
	WakeCount      *int     `json:"wake_count"`
	FoodName       *string  `json:"food_name"`
	FoodType       *string  `json:"food_type"`
	AmountLevel    *int     `json:"amount_level"`
	RecordTime     *string  `json:"record_time"`
	SupplementName *string  `json:"supplement_name"`
	TakenAt        *string  `json:"taken_at"`
	Type           *string  `json:"type"`
	Value          *float64 `json:"value"`
	MeasuredAt     *string  `json:"measured_at"`
}

// ParsedResult represents the full LLM output.
type ParsedResult struct {
	Intent  string         `json:"intent"`
	Records []ParsedRecord `json:"records"`
	Error   *string        `json:"error"`
	Query   *ParsedQuery   `json:"query"`
}

// ParsedQuery represents a query intent from LLM.
type ParsedQuery struct {
	Category  string `json:"category"`
	TimeRange string `json:"time_range"`
}

func (b *Bot) handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	log.Printf("[DEBUG][feishu] handleMessage called")

	// Only handle text messages
	msgType := event.Event.Message.MessageType
	if msgType == nil || *msgType != "text" {
		log.Printf("[DEBUG][feishu] skipping non-text message, type: %v", msgType)
		return nil
	}

	// Extract text content
	content := event.Event.Message.Content
	if content == nil {
		log.Printf("[DEBUG][feishu] content is nil")
		return nil
	}

	log.Printf("[DEBUG][feishu] raw content json: %s", *content)

	var textMsg struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(*content), &textMsg); err != nil {
		log.Printf("[DEBUG][feishu] parse message content error: %v", err)
		return nil
	}

	userText := strings.TrimSpace(textMsg.Text)
	if userText == "" {
		log.Printf("[DEBUG][feishu] userText is empty")
		return nil
	}

	log.Printf("[DEBUG][feishu] received message from user: %s", userText)

	// Call LLM to parse
	replyText := b.processMessage(userText)

	log.Printf("[DEBUG][feishu] reply text: %s", replyText)

	// Send reply
	b.replyMessage(ctx, event, replyText)

	return nil
}

func (b *Bot) processMessage(userText string) string {
	log.Printf("[DEBUG][feishu] processMessage called with text: %s", userText)
	parsed, err := b.callLLM(userText)
	if err != nil {
		log.Printf("[DEBUG][feishu] LLM error: %v", err)
		return "❌ 解析失败，请重新描述"
	}

	log.Printf("[DEBUG][feishu] LLM parsed result: %+v", parsed)

	// Query intent
	if parsed.Intent == "query" && parsed.Query != nil {
		result, err := b.ExecuteQuery(parsed.Query.Category, parsed.Query.TimeRange)
		if err != nil {
			log.Printf("[DEBUG][feishu] query failed: %v", err)
			return fmt.Sprintf("❌ 查询失败: %s", err.Error())
		}
		return result.Text
	}

	// Record intent
	if len(parsed.Records) == 0 {
		if parsed.Error != nil {
			return fmt.Sprintf("🤔 %s，请换种说法试试", *parsed.Error)
		}
		return "🤔 没有识别到任何记录，请换种说法试试"
	}

	var results []string
	for _, record := range parsed.Records {
		log.Printf("[DEBUG][feishu] saving record: %+v", record)
		result, err := b.saveRecord(record)
		if err != nil {
			log.Printf("[DEBUG][feishu] save record failed: %v", err)
			results = append(results, fmt.Sprintf("❌ 保存失败: %s", err.Error()))
			continue
		}
		log.Printf("[DEBUG][feishu] saved record successfully: %s", result)
		results = append(results, result)
	}

	if len(results) == 0 {
		return "❌ 所有记录都保存失败了"
	}

	return "✅ 已记录：\n" + strings.Join(results, "\n")
}

func (b *Bot) callLLM(userText string) (*ParsedResult, error) {
	systemPrompt := SystemPrompt()
	log.Printf("[DEBUG][feishu] callLLM called, system prompt length: %d, user text: %s", len(systemPrompt), userText)
	reply, err := b.llm.Chat(systemPrompt, userText)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG][feishu] LLM raw reply: %s", reply)

	// Extract JSON from response (might have markdown code blocks)
	content := strings.TrimSpace(reply)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		if len(lines) > 2 {
			content = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	log.Printf("[DEBUG][feishu] LLM extracted json content: %s", content)

	var result ParsedResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse LLM output: %w, raw: %s", err, content)
	}

	log.Printf("[DEBUG][feishu] LLM parsed %d records", len(result.Records))

	return &result, nil
}

func (b *Bot) getFirstChildID() (string, error) {
	children, err := b.childStore.GetAll()
	if err != nil {
		return "", err
	}
	if len(children) == 0 {
		return "", fmt.Errorf("没有孩子")
	}
	return children[0].ID, nil
}

func (b *Bot) saveRecord(record ParsedRecord) (string, error) {
	switch record.Category {
	case "sleep":
		return b.saveSleep(record)
	case "diet":
		return b.saveDiet(record)
	case "supplement":
		return b.saveSupplement(record)
	case "measurement":
		return b.saveMeasurement(record)
	default:
		return "", fmt.Errorf("未知记录类型: %s", record.Category)
	}
}

func (b *Bot) saveSleep(r ParsedRecord) (string, error) {
	if r.StartTime == nil {
		return "", fmt.Errorf("缺少开始时间")
	}

	childID, err := b.getFirstChildID()
	if err != nil {
		return "", fmt.Errorf("获取孩子信息: %w", err)
	}

	startTime, err := parseTime(*r.StartTime)
	if err != nil {
		return "", fmt.Errorf("时间格式错误: %s", *r.StartTime)
	}

	var endTime *time.Time
	if r.EndTime != nil {
		t, err := parseTime(*r.EndTime)
		if err != nil {
			return "", fmt.Errorf("结束时间格式错误: %s", *r.EndTime)
		}
		endTime = &t
	}

	wokeUp := false
	if r.WokeUp != nil {
		wokeUp = *r.WokeUp
	}
	wakeCount := 0
	if r.WakeCount != nil {
		wakeCount = *r.WakeCount
	}

	rec := &model.SleepRecord{
		ID:        uuid.NewString(),
		ChildID:   childID,
		StartTime: startTime,
		EndTime:   endTime,
		WokeUp:    wokeUp,
		WakeCount: wakeCount,
		CreatedBy: "feishu-bot",
		CreatedAt: time.Now(),
	}

	if err := b.sleepStore.Create(rec); err != nil {
		return "", fmt.Errorf("保存睡眠记录: %w", err)
	}

	label := formatTime(startTime)
	if endTime != nil {
		diff := endTime.Sub(startTime).Minutes()
		h := int(diff) / 60
		m := int(diff) % 60
		label = fmt.Sprintf("%s - %s（%d小时%d分钟）", formatTime(startTime), formatTime(*endTime), h, m)
	} else {
		label = fmt.Sprintf("%s - 进行中", formatTime(startTime))
	}
	if wokeUp {
		label += fmt.Sprintf("，中途醒来%d次", wakeCount)
	}
	return "睡眠：" + label, nil
}

func (b *Bot) saveDiet(r ParsedRecord) (string, error) {
	if r.FoodName == nil {
		return "", fmt.Errorf("缺少食物名称")
	}

	childID, err := b.getFirstChildID()
	if err != nil {
		return "", fmt.Errorf("获取孩子信息: %w", err)
	}

	amountLevel := 2
	if r.AmountLevel != nil {
		amountLevel = *r.AmountLevel
	}

	recordTime := time.Now()
	if r.RecordTime != nil {
		t, err := parseTime(*r.RecordTime)
		if err == nil {
			recordTime = t
		}
	}

	foodType := "staple"
	if r.FoodType != nil {
		foodType = *r.FoodType
	}

	mealType := ""
	if r.MealType != nil {
		mealType = *r.MealType
	}

	rec := &model.DietRecord{
		ID:          uuid.NewString(),
		ChildID:     childID,
		FoodName:    *r.FoodName,
		FoodType:    foodType,
		AmountLevel: amountLevel,
		RecordTime:  recordTime,
		MealGroupID: r.MealGroupID,
		MealType:    mealType,
		CreatedBy:   "feishu-bot",
		CreatedAt:   time.Now(),
	}

	if err := b.dietStore.Create(rec); err != nil {
		return "", fmt.Errorf("保存饮食记录: %w", err)
	}

	amountLabels := map[int]string{1: "少", 2: "正常", 3: "多"}
	return fmt.Sprintf("饮食：%s（%s，%s）", *r.FoodName, foodType, amountLabels[amountLevel]), nil
}

func (b *Bot) saveSupplement(r ParsedRecord) (string, error) {
	if r.SupplementName == nil {
		return "", fmt.Errorf("缺少补剂名称")
	}

	childID, err := b.getFirstChildID()
	if err != nil {
		return "", fmt.Errorf("获取孩子信息: %w", err)
	}

	takenAt := time.Now()
	if r.TakenAt != nil {
		t, err := parseTime(*r.TakenAt)
		if err == nil {
			takenAt = t
		}
	}

	rec := &model.SupplementRecord{
		ID:             uuid.NewString(),
		ChildID:        childID,
		SupplementName: *r.SupplementName,
		TakenAt:        takenAt,
		CreatedBy:      "feishu-bot",
		CreatedAt:      time.Now(),
	}

	if err := b.suppStore.Create(rec); err != nil {
		return "", fmt.Errorf("保存补剂记录: %w", err)
	}

	return fmt.Sprintf("补剂：%s（%s）", *r.SupplementName, formatTime(takenAt)), nil
}

func (b *Bot) saveMeasurement(r ParsedRecord) (string, error) {
	if r.Type == nil || r.Value == nil {
		return "", fmt.Errorf("缺少测量类型或数值")
	}

	childID, err := b.getFirstChildID()
	if err != nil {
		return "", fmt.Errorf("获取孩子信息: %w", err)
	}

	measuredAt := time.Now()
	if r.MeasuredAt != nil {
		t, err := parseTime(*r.MeasuredAt)
		if err == nil {
			measuredAt = t
		}
	}

	unit := "cm"
	if *r.Type == "weight" {
		unit = "kg"
	}

	rec := &model.Measurement{
		ID:         uuid.NewString(),
		ChildID:    childID,
		Type:       *r.Type,
		Value:      *r.Value,
		MeasuredAt: measuredAt,
		CreatedBy:  "feishu-bot",
		CreatedAt:  time.Now(),
	}

	if err := b.measureStore.Create(rec); err != nil {
		return "", fmt.Errorf("保存测量记录: %w", err)
	}

	typeLabels := map[string]string{
		"weight": "体重",
		"height": "身高",
	}
	return fmt.Sprintf("测量：%s %.1f%s", typeLabels[*r.Type], *r.Value, unit), nil
}

func parseTime(s string) (time.Time, error) {
	// Try RFC3339 first (with timezone)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Try without timezone (e.g. "2026-05-08T11:00:00")
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}
	// Try date only (e.g. "2026-05-08")
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", s)
}

func formatTime(t time.Time) string {
	return t.Format("15:04")
}

func (b *Bot) replyMessage(ctx context.Context, event *larkim.P2MessageReceiveV1, text string) {
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(*event.Event.Message.MessageId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			Content(fmt.Sprintf(`{"text":%q}`, text)).
			Build()).
		Build()

	_, err := b.larkClient.Im.Message.Reply(ctx, req)
	if err != nil {
		log.Printf("feishu: reply message failed: %v", err)
	}
}
