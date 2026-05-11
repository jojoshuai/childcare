package feishu

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// QueryResult holds the formatted reply for a query intent.
type QueryResult struct {
	Category string
	Text     string
}

// ExecuteQuery executes a query against the stores and returns formatted text.
func (b *Bot) ExecuteQuery(category, timeRange string) (*QueryResult, error) {
	childID, err := b.getFirstChildID()
	if err != nil {
		return nil, fmt.Errorf("获取孩子信息: %w", err)
	}

	start, end := timeRangeToRange(timeRange)

	var text string
	switch category {
	case "sleep":
		text = b.querySleep(childID, start, end)
	case "diet":
		text = b.queryDiet(childID, start, end)
	case "supplement":
		text = b.querySupplement(childID, start, end)
	case "summary":
		sleep := b.querySleep(childID, start, end)
		diet := b.queryDiet(childID, start, end)
		supp := b.querySupplement(childID, start, end)
		text = fmt.Sprintf("%s\n%s\n%s", sleep, diet, supp)
	default:
		return nil, fmt.Errorf("未知查询类型: %s", category)
	}

	return &QueryResult{Category: category, Text: text}, nil
}

func timeRangeToRange(tr string) (start, end time.Time) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch tr {
	case "today":
		return today, today.Add(24 * time.Hour)
	case "yesterday":
		y := today.Add(-24 * time.Hour)
		return y, today
	case "last_7_days":
		return today.Add(-7 * 24 * time.Hour), today.Add(24 * time.Hour)
	default:
		return today, today.Add(24 * time.Hour)
	}
}

func (b *Bot) querySleep(childID string, start, end time.Time) string {
	records, err := b.sleepStore.GetByChildID(childID)
	if err != nil || len(records) == 0 {
		return "睡眠：暂无记录"
	}

	var totalMinutes float64
	var totalWakeCount int
	var count int

	for _, r := range records {
		if r.StartTime.Before(start) || !r.StartTime.Before(end) {
			continue
		}
		count++
		if r.EndTime != nil {
			totalMinutes += r.EndTime.Sub(r.StartTime).Minutes()
		}
		totalWakeCount += r.WakeCount
	}

	if count == 0 {
		return "睡眠：暂无记录"
	}

	totalHours := totalMinutes / 60.0
	h := int(totalHours)
	m := int(math.Mod(totalMinutes, 60))

	label := timeRangeLabel(start, end)
	parts := []string{fmt.Sprintf("睡眠（%s）：共 %d 小时 %d 分钟", label, h, m)}
	if count > 1 {
		parts[0] += fmt.Sprintf("（%d 条记录）", count)
	}
	if totalWakeCount > 0 {
		parts = append(parts, fmt.Sprintf("共醒来 %d 次", totalWakeCount))
	}

	return strings.Join(parts, "，")
}

func (b *Bot) queryDiet(childID string, start, end time.Time) string {
	records, err := b.dietStore.GetByChildID(childID)
	if err != nil || len(records) == 0 {
		return "饮食：暂无记录"
	}

	// Group by meal_group_id to count meals
	groups := make(map[string]int)
	totalItems := 0
	for _, r := range records {
		if r.RecordTime.Before(start) || !r.RecordTime.Before(end) {
			continue
		}
		gid := ""
		if r.MealGroupID != nil {
			gid = *r.MealGroupID
		}
		groups[gid]++
		totalItems++
	}

	mealCount := len(groups)
	if mealCount == 0 {
		return "饮食：暂无记录"
	}

	label := timeRangeLabel(start, end)
	return fmt.Sprintf("饮食（%s）：共 %d 餐，%d 种食物", label, mealCount, totalItems)
}

func (b *Bot) querySupplement(childID string, start, end time.Time) string {
	records, err := b.suppStore.GetByChildID(childID)
	if err != nil || len(records) == 0 {
		return "补剂：暂无记录"
	}

	typeCount := make(map[string]int)
	for _, r := range records {
		if r.TakenAt.Before(start) || !r.TakenAt.Before(end) {
			continue
		}
		typeCount[r.SupplementName]++
	}

	if len(typeCount) == 0 {
		return "补剂：暂无记录"
	}

	label := timeRangeLabel(start, end)
	parts := []string{fmt.Sprintf("补剂（%s）", label)}
	for name, count := range typeCount {
		parts = append(parts, fmt.Sprintf("%s %d 次", name, count))
	}

	return strings.Join(parts, "，")
}

func timeRangeLabel(start, end time.Time) string {
	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	yesterday := today.Add(-24 * time.Hour)

	if start.Equal(today) && end.Equal(today.Add(24*time.Hour)) {
		return "今天"
	}
	if start.Equal(yesterday) && end.Equal(today) {
		return "昨天"
	}
	return fmt.Sprintf("%s 至 %s", start.Format("01-02"), end.Format("01-02"))
}
