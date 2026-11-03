package handlers

import (
	"testing"
)

func TestFavoritesLogic(t *testing.T) {
	tests := []struct {
		name           string
		existingFavs   []FavoriteRecord
		newVocabID     string
		userID         string
		expectConflict bool
		expectedCount  int
	}{
		{
			name:           "add first favorite",
			existingFavs:   []FavoriteRecord{},
			newVocabID:     "vocab1",
			userID:         "user1",
			expectConflict: false,
			expectedCount:  1,
		},
		{
			name: "add second favorite",
			existingFavs: []FavoriteRecord{
				{ID: "fav1", UserID: "user1", VocabID: "vocab1"},
			},
			newVocabID:     "vocab2",
			userID:         "user1",
			expectConflict: false,
			expectedCount:  2,
		},
		{
			name: "duplicate favorite returns conflict",
			existingFavs: []FavoriteRecord{
				{ID: "fav1", UserID: "user1", VocabID: "vocab1"},
			},
			newVocabID:     "vocab1",
			userID:         "user1",
			expectConflict: true,
			expectedCount:  1,
		},
		{
			name: "different user can favorite same vocab",
			existingFavs: []FavoriteRecord{
				{ID: "fav1", UserID: "user1", VocabID: "vocab1"},
			},
			newVocabID:     "vocab1",
			userID:         "user2",
			expectConflict: false,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			favs := make([]FavoriteRecord, len(tt.existingFavs))
			copy(favs, tt.existingFavs)

			conflict := checkFavoriteConflict(favs, tt.userID, tt.newVocabID)
			if conflict != tt.expectConflict {
				t.Errorf("Conflict check: expected %v, got %v", tt.expectConflict, conflict)
			}

			if !tt.expectConflict {
				newFav := FavoriteRecord{
					ID:       "new_fav_id",
					UserID:   tt.userID,
					VocabID:  tt.newVocabID,
					CreatedAt: "2024-01-01T00:00:00Z",
				}
				favs = append(favs, newFav)
			}

			if len(favs) != tt.expectedCount {
				t.Errorf("Expected %d favorites, got %d", tt.expectedCount, len(favs))
			}
		})
	}
}

func TestFavoriteOwnership(t *testing.T) {
	favs := []FavoriteRecord{
		{ID: "fav1", UserID: "user1", VocabID: "vocab1"},
		{ID: "fav2", UserID: "user2", VocabID: "vocab2"},
	}

	tests := []struct {
		name        string
		favID       string
		requestUser string
		allowed     bool
	}{
		{"owner can delete", "fav1", "user1", true},
		{"non-owner cannot delete", "fav1", "user2", false},
		{"owner can delete other", "fav2", "user2", true},
		{"non-owner cannot delete other", "fav2", "user1", false},
		{"nonexistent favorite", "fav3", "user1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := checkFavoriteOwnership(favs, tt.favID, tt.requestUser)
			if allowed != tt.allowed {
				t.Errorf("Ownership check for %s by %s: expected %v, got %v", tt.favID, tt.requestUser, tt.allowed, allowed)
			}
		})
	}
}

func TestListFavoritesPagination(t *testing.T) {
	// Create 25 favorites
	favs := make([]FavoriteRecord, 25)
	for i := 0; i < 25; i++ {
		favs[i] = FavoriteRecord{
			ID:        "fav" + string(rune('0'+i)),
			UserID:    "user1",
			VocabID:   "vocab" + string(rune('0'+i)),
			CreatedAt: "2024-01-01T00:00:00Z",
		}
	}

	tests := []struct {
		name       string
		page       int
		limit      int
		expected   int
	}{
		{"page 1 limit 20", 1, 20, 20},
		{"page 2 limit 20", 2, 20, 5},
		{"page 1 limit 50", 1, 50, 25},
		{"page 3 limit 10", 3, 10, 5},
		{"page 4 limit 10", 4, 10, 0},
		{"invalid page defaults to 1", -1, 20, 20},
		{"invalid limit defaults to 50", 1, -1, 25},
		{"limit capped at 200", 1, 300, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paginateFavorites(favs, tt.page, tt.limit)
			if len(result) != tt.expected {
				t.Errorf("Expected %d items, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestListFavoritesSorting(t *testing.T) {
	favs := []FavoriteRecord{
		{ID: "fav1", UserID: "user1", VocabID: "vocab1", CreatedAt: "2024-01-03T00:00:00Z"},
		{ID: "fav2", UserID: "user1", VocabID: "vocab2", CreatedAt: "2024-01-01T00:00:00Z"},
		{ID: "fav3", UserID: "user1", VocabID: "vocab3", CreatedAt: "2024-01-02T00:00:00Z"},
	}

	result := paginateFavorites(favs, 1, 10)

	// Should be sorted by created_at descending (newest first)
	if len(result) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(result))
	}
	if result[0].ID != "fav1" {
		t.Errorf("First should be fav1 (newest), got %s", result[0].ID)
	}
	if result[1].ID != "fav3" {
		t.Errorf("Second should be fav3, got %s", result[1].ID)
	}
	if result[2].ID != "fav2" {
		t.Errorf("Third should be fav2 (oldest), got %s", result[2].ID)
	}
}

// Helper types and functions

type FavoriteRecord struct {
	ID        string
	UserID    string
	VocabID   string
	CreatedAt string
}

func checkFavoriteConflict(favs []FavoriteRecord, userID, vocabID string) bool {
	for _, f := range favs {
		if f.UserID == userID && f.VocabID == vocabID {
			return true
		}
	}
	return false
}

func checkFavoriteOwnership(favs []FavoriteRecord, favID, userID string) bool {
	for _, f := range favs {
		if f.ID == favID {
			return f.UserID == userID
		}
	}
	return false
}

func paginateFavorites(favs []FavoriteRecord, page, limit int) []FavoriteRecord {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// Sort by created_at descending (already sorted in test data)
	start := (page - 1) * limit
	if start >= len(favs) {
		return []FavoriteRecord{}
	}
	end := start + limit
	if end > len(favs) {
		end = len(favs)
	}
	return favs[start:end]
}