package handlers

import (
	"testing"
	"time"
)

func TestProgressStatsCalculation(t *testing.T) {
	// Test the logic for calculating progress stats
	tests := []struct {
		name           string
		records        []ProgressRecord
		expectedStats  map[string]int
	}{
		{
			name: "empty records",
			records: []ProgressRecord{},
			expectedStats: map[string]int{
				"total_words":      0,
				"level_0_new":      0,
				"level_1_learning": 0,
				"level_2_familiar": 0,
				"level_3_known":    0,
				"level_4_mastered": 0,
				"level_5_fluent":   0,
				"total_reviews":    0,
			},
		},
		{
			name: "mixed levels",
			records: []ProgressRecord{
				{Level: 0, Reviews: 0},
				{Level: 0, Reviews: 1},
				{Level: 1, Reviews: 2},
				{Level: 2, Reviews: 3},
				{Level: 3, Reviews: 4},
				{Level: 4, Reviews: 5},
				{Level: 5, Reviews: 6},
			},
			expectedStats: map[string]int{
				"total_words":      7,
				"level_0_new":      2,
				"level_1_learning": 1,
				"level_2_familiar": 1,
				"level_3_known":    1,
				"level_4_mastered": 1,
				"level_5_fluent":   1,
				"total_reviews":    21,
			},
		},
		{
			name: "all level 5",
			records: []ProgressRecord{
				{Level: 5, Reviews: 10},
				{Level: 5, Reviews: 20},
				{Level: 5, Reviews: 30},
			},
			expectedStats: map[string]int{
				"total_words":      3,
				"level_0_new":      0,
				"level_1_learning": 0,
				"level_2_familiar": 0,
				"level_3_known":    0,
				"level_4_mastered": 0,
				"level_5_fluent":   3,
				"total_reviews":    60,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := calculateProgressStats(tt.records)

			for key, expected := range tt.expectedStats {
				if stats[key] != expected {
					t.Errorf("Stat %s: expected %d, got %d", key, expected, stats[key])
				}
			}
		})
	}
}

func TestSpacedRepetitionIntervals(t *testing.T) {
	// Test the spaced repetition logic for review intervals
	tests := []struct {
		name     string
		level    int
		expected string // description of expected interval
	}{
		{name: "level 0 - daily", level: 0, expected: "daily"},
		{name: "level 1 - daily", level: 1, expected: "daily"},
		{name: "level 2 - every 3 days", level: 2, expected: "3 days"},
		{name: "level 3 - weekly", level: 3, expected: "weekly"},
		{name: "level 4 - biweekly", level: 4, expected: "biweekly"},
		{name: "level 5 - monthly", level: 5, expected: "monthly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := getReviewInterval(tt.level)
			// Just verify the function returns something reasonable
			if interval <= 0 {
				t.Errorf("Level %d should have positive interval", tt.level)
			}
			// Higher levels should have longer intervals
			if tt.level > 0 {
				prevInterval := getReviewInterval(tt.level - 1)
				if interval <= prevInterval {
					t.Errorf("Level %d interval (%d) should be > level %d interval (%d)", tt.level, interval, tt.level-1, prevInterval)
				}
			}
		})
	}
}

func TestKnowledgeLevelBounds(t *testing.T) {
	// Test that knowledge level stays within 0-5 bounds
	tests := []struct {
		name           string
		currentLevel   int
		correct        bool
		expectedLevel  int
	}{
		{"level 0 correct -> 1", 0, true, 1},
		{"level 1 correct -> 2", 1, true, 2},
		{"level 4 correct -> 5", 4, true, 5},
		{"level 5 correct -> 5 (max)", 5, true, 5},
		{"level 1 incorrect -> 0", 1, false, 0},
		{"level 2 incorrect -> 1", 2, false, 1},
		{"level 5 incorrect -> 4", 5, false, 4},
		{"level 0 incorrect -> 0 (min)", 0, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLevel := adjustKnowledgeLevel(tt.currentLevel, tt.correct)
			if newLevel != tt.expectedLevel {
				t.Errorf("adjustKnowledgeLevel(%d, %v) = %d, want %d", tt.currentLevel, tt.correct, newLevel, tt.expectedLevel)
			}
		})
	}
}

func TestProgressRecordUpdate(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	tests := []struct {
		name          string
		existing      *ProgressRecord
		newLevel      int
		expectNew     bool
		expectedLevel int
		expectedReviews int
	}{
		{
			name:          "create new record",
			existing:      nil,
			newLevel:      2,
			expectNew:     true,
			expectedLevel: 2,
			expectedReviews: 1,
		},
		{
			name: "update existing record correct",
			existing: &ProgressRecord{
				Level:    1,
				Reviews:  3,
				Reviewed: now,
			},
			newLevel:       2,
			expectNew:      false,
			expectedLevel:  2,
			expectedReviews: 4,
		},
		{
			name: "update existing record incorrect",
			existing: &ProgressRecord{
				Level:    3,
				Reviews:  5,
				Reviewed: now,
			},
			newLevel:       2,
			expectNew:      false,
			expectedLevel:  2,
			expectedReviews: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := updateProgressRecord(tt.existing, tt.newLevel, now)

			if record.Level != tt.expectedLevel {
				t.Errorf("Level: expected %d, got %d", tt.expectedLevel, record.Level)
			}
			if record.Reviews != tt.expectedReviews {
				t.Errorf("Reviews: expected %d, got %d", tt.expectedReviews, record.Reviews)
			}
			if record.Reviewed != now {
				t.Errorf("Reviewed timestamp not updated")
			}
		})
	}
}

// Helper types and functions for testing (extracted from handler logic)

type ProgressRecord struct {
	Level    int
	Reviews  int
	Reviewed string
}

func calculateProgressStats(records []ProgressRecord) map[string]int {
	stats := map[string]int{
		"total_words":      len(records),
		"level_0_new":      0,
		"level_1_learning": 0,
		"level_2_familiar": 0,
		"level_3_known":    0,
		"level_4_mastered": 0,
		"level_5_fluent":   0,
		"total_reviews":    0,
	}

	for _, rec := range records {
		stats["total_reviews"] += rec.Reviews
		switch rec.Level {
		case 0:
			stats["level_0_new"]++
		case 1:
			stats["level_1_learning"]++
		case 2:
			stats["level_2_familiar"]++
		case 3:
			stats["level_3_known"]++
		case 4:
			stats["level_4_mastered"]++
		case 5:
			stats["level_5_fluent"]++
		}
	}

	return stats
}

func getReviewInterval(level int) int {
	switch level {
	case 0, 1:
		return 1 // daily
	case 2:
		return 3 // 3 days
	case 3:
		return 7 // weekly
	case 4:
		return 14 // biweekly
	case 5:
		return 30 // monthly
	default:
		return 1
	}
}

func adjustKnowledgeLevel(currentLevel int, correct bool) int {
	if correct {
		if currentLevel < 5 {
			return currentLevel + 1
		}
		return 5
	} else {
		if currentLevel > 0 {
			return currentLevel - 1
		}
		return 0
	}
}

func updateProgressRecord(existing *ProgressRecord, newLevel int, now string) *ProgressRecord {
	if existing == nil {
		return &ProgressRecord{
			Level:    newLevel,
			Reviews:  1,
			Reviewed: now,
		}
	}

	return &ProgressRecord{
		Level:    newLevel,
		Reviews:  existing.Reviews + 1,
		Reviewed: now,
	}
}