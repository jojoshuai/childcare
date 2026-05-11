// +build ignore

package main

import (
	"fmt"
	"math/rand"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "childcare:childcare123@tcp(localhost:3306)/childcare?parseTime=true&charset=utf8mb4&multiStatements=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("=== Clearing existing data ===")

	// Clear in order (foreign key constraints)
	tables := []string{
		"sleep_records",
		"diet_records",
		"supplement_records",
		"measurements",
	}
	for _, t := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", t))
		if err != nil {
			panic(fmt.Sprintf("clear %s: %v", t, err))
		}
		fmt.Printf("  cleared %s\n", t)
	}

	// Get the child
	var childID, childName string
	err = db.QueryRow("SELECT id, name FROM children LIMIT 1").Scan(&childID, &childName)
	if err != nil {
		panic(fmt.Sprintf("get child: %v", err))
	}
	fmt.Printf("  using child: %s (%s)\n", childName, childID)

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -35)
	rng := rand.New(rand.NewSource(42))

	fmt.Println("\n=== Generating 35 days of data ===")

	// Sleep records: ~1 night sleep + 1-2 naps per day
	sleepCount := 0
	for d := thirtyDaysAgo; d.Before(now); d = d.AddDate(0, 0, 1) {
		// Night sleep: 20:00-21:30 -> 06:30-08:00 next day
		sleepHour := 20 + rng.Intn(2) // 20 or 21
		sleepMin := rng.Intn(60)
		startTime := time.Date(d.Year(), d.Month(), d.Day(), sleepHour, sleepMin, 0, 0, d.Location())

		wakeHour := 6 + rng.Intn(3) // 6-8
		wakeMin := rng.Intn(60)
		endTime := startTime.AddDate(0, 0, 1)
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), wakeHour, wakeMin, 0, 0, endTime.Location())

		wokeUp := rng.Intn(3) > 0 // 2/3 chance
		wakeCount := 0
		if wokeUp {
			wakeCount = 1 + rng.Intn(3) // 1-3 times
		}

		_, err := db.Exec(
			`INSERT INTO sleep_records (id, child_id, start_time, end_time, woke_up, wake_count, created_by, created_at)
			 VALUES (UUID(), ?, ?, ?, ?, ?, 'seed', ?)`,
			childID, startTime, endTime, wokeUp, wakeCount, now,
		)
		if err != nil {
			panic(fmt.Sprintf("insert sleep: %v", err))
		}
		sleepCount++

		// Daytime naps: 1-2 per day
		napCount := 1 + rng.Intn(2)
		napStarts := []int{12, 14} // noon nap, afternoon nap
		for i := 0; i < napCount && i < len(napStarts); i++ {
			nsHour := napStarts[i] + rng.Intn(1)
			nsMin := rng.Intn(60)
			nsTime := time.Date(d.Year(), d.Month(), d.Day(), nsHour, nsMin, 0, 0, d.Location())
			neTime := nsTime.Add(30*time.Minute + time.Duration(rng.Intn(60))*time.Minute)

			_, err := db.Exec(
				`INSERT INTO sleep_records (id, child_id, start_time, end_time, woke_up, wake_count, created_by, created_at)
				 VALUES (UUID(), ?, ?, ?, 0, 0, 'seed', ?)`,
				childID, nsTime, neTime, now,
			)
			if err != nil {
				panic(fmt.Sprintf("insert nap sleep: %v", err))
			}
			sleepCount++
		}
	}
	fmt.Printf("  generated %d sleep records\n", sleepCount)

	// Diet records: 3 meals + 1-2 snacks per day, grouped by meal
	dietCount := 0
	foods := map[string][]string{
		"breakfast": {"米粉", "粥", "鸡蛋", "面包", "牛奶", "燕麦", "包子", "豆浆", "面条"},
		"lunch":     {"米饭", "面条", "猪肉", "鸡肉", "牛肉", "鱼肉", "西兰花", "胡萝卜", "土豆", "豆腐", "西红柿"},
		"dinner":    {"米饭", "粥", "馒头", "鸡蛋", "虾", "猪肉", "白菜", "菠菜", "南瓜", "红薯"},
		"snack":     {"小馒头", "水果泥", "酸奶", "饼干", "苹果", "香蕉", "梨", "蓝莓"},
	}
	mealHours := map[string][]int{
		"breakfast": {7, 8},
		"lunch":     {11, 12},
		"dinner":    {17, 18},
		"snack":     {9, 15, 16},
	}

	for d := thirtyDaysAgo; d.Before(now); d = d.AddDate(0, 0, 1) {
		mealTypes := []string{"breakfast", "lunch", "dinner"}
		// Random snack
		if rng.Intn(3) > 0 {
			mealTypes = append(mealTypes, "snack")
		}

		for _, mealType := range mealTypes {
			mealGroupID := fmt.Sprintf("seed-%s-%s", d.Format("2006-01-02"), mealType)
			hours := mealHours[mealType]
			h := hours[rng.Intn(len(hours))]
			m := rng.Intn(60)

			// 1-3 items per meal
			itemCount := 1 + rng.Intn(3)
			foodList := foods[mealType]

			for i := 0; i < itemCount; i++ {
				foodName := foodList[rng.Intn(len(foodList))]
				foodType := foodTypeForMeal(mealType, foodName, rng)
				amountLevel := 1 + rng.Intn(3)
				recordTime := time.Date(d.Year(), d.Month(), d.Day(), h, m+rng.Intn(10), 0, 0, d.Location())

				_, err := db.Exec(
					`INSERT INTO diet_records (id, child_id, food_name, food_type, amount_level, record_time, meal_group_id, meal_type, created_by, created_at)
					 VALUES (UUID(), ?, ?, ?, ?, ?, ?, ?, 'seed', ?)`,
					childID, foodName, foodType, amountLevel, recordTime, mealGroupID, mealType, now,
				)
				if err != nil {
					panic(fmt.Sprintf("insert diet: %v", err))
				}
				dietCount++
			}
		}
	}
	fmt.Printf("  generated %d diet records\n", dietCount)

	// Supplement records: 1-2 per day
	supplements := []string{"维生素D", "钙", "DHA", "维生素AD", "锌", "铁"}
	suppCount := 0
	for d := thirtyDaysAgo; d.Before(now); d = d.AddDate(0, 0, 1) {
		// Vitamin D daily (most common)
		suppTime := time.Date(d.Year(), d.Month(), d.Day(), 8+rng.Intn(3), rng.Intn(60), 0, 0, d.Location())
		_, err := db.Exec(
			`INSERT INTO supplement_records (id, child_id, supplement_name, taken_at, created_by, created_at)
			 VALUES (UUID(), ?, '维生素D', ?, 'seed', ?)`,
			childID, suppTime, now,
		)
		if err != nil {
			panic(fmt.Sprintf("insert supplement: %v", err))
		}
		suppCount++

		// Random second supplement
		if rng.Intn(2) == 0 {
			supp := supplements[1+rng.Intn(len(supplements)-1)]
			suppTime2 := time.Date(d.Year(), d.Month(), d.Day(), 12+rng.Intn(4), rng.Intn(60), 0, 0, d.Location())
			_, err := db.Exec(
				`INSERT INTO supplement_records (id, child_id, supplement_name, taken_at, created_by, created_at)
				 VALUES (UUID(), ?, ?, ?, 'seed', ?)`,
				childID, supp, suppTime2, now,
			)
			if err != nil {
				panic(fmt.Sprintf("insert supplement: %v", err))
			}
			suppCount++
		}

		// Skip a random day (simulate missed supplements)
		if rng.Intn(10) == 0 {
			d = d.AddDate(0, 0, 1)
		}
	}
	fmt.Printf("  generated %d supplement records\n", suppCount)

	// Measurement records: ~1 per week
	measureCount := 0
	for d := thirtyDaysAgo; d.Before(now); d = d.AddDate(0, 0, 7) {
		types := []string{"weight", "height"}
		for _, t := range types {
			var value float64
			switch t {
			case "weight":
				value = 7.5 + rng.Float64()*3.0
			case "height":
				value = 65.0 + rng.Float64()*15.0
			}
			mTime := time.Date(d.Year(), d.Month(), d.Day(), 10, rng.Intn(60), 0, 0, d.Location())
			_, err := db.Exec(
				`INSERT INTO measurements (id, child_id, type, value, measured_at, created_by, created_at)
				 VALUES (UUID(), ?, ?, ?, ?, 'seed', ?)`,
				childID, t, value, mTime, now,
			)
			if err != nil {
				panic(fmt.Sprintf("insert measurement: %v", err))
			}
			measureCount++
		}
	}
	fmt.Printf("  generated %d measurement records\n", measureCount)

	fmt.Println("\n=== Done ===")
}

func foodTypeForMeal(mealType, foodName string, rng *rand.Rand) string {
	// Map food names to types
	typeMap := map[string]string{
		"米粉": "staple", "粥": "staple", "面包": "staple", "燕麦": "staple", "包子": "staple", "豆浆": "dairy",
		"面条": "staple", "米饭": "staple", "馒头": "staple", "红薯": "staple", "土豆": "staple", "豆腐": "protein",
		"鸡蛋": "protein", "猪肉": "protein", "鸡肉": "protein", "牛肉": "protein", "鱼肉": "protein", "虾": "protein",
		"西兰花": "vegetable", "胡萝卜": "vegetable", "白菜": "vegetable", "菠菜": "vegetable", "西红柿": "vegetable",
		"南瓜": "vegetable", "青椒": "vegetable",
		"牛奶": "dairy", "酸奶": "dairy",
		"小馒头": "snack", "水果泥": "fruit", "饼干": "snack", "苹果": "fruit", "香蕉": "fruit",
		"梨": "fruit", "蓝莓": "fruit",
	}
	if t, ok := typeMap[foodName]; ok {
		return t
	}
	types := []string{"staple", "vegetable", "fruit", "protein", "dairy", "snack"}
	return types[rng.Intn(len(types))]
}
