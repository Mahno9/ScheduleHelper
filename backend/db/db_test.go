package db_test

import (
	"os"
	"testing"
	"time"

	"schedulehelper/backend/db"
	"schedulehelper/backend/models"
)

func setupDB(t *testing.T) *db.DB {
	t.Helper()
	f, err := os.CreateTemp("", "schedulehelper-test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	d, err := db.Open(f.Name())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestCreateAndGetUser(t *testing.T) {
	d := setupDB(t)

	req := models.RegisterRequest{
		Name:  "Alice",
		Emoji: "🐱",
		Color: "#FF5733",
	}
	user, err := d.CreateUser(req)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}
	if user.Name != "Alice" {
		t.Errorf("expected name Alice, got %s", user.Name)
	}

	got, err := d.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("got name %s, want Alice", got.Name)
	}
}

func TestUserNameUniqueness(t *testing.T) {
	d := setupDB(t)
	req := models.RegisterRequest{Name: "Bob", Emoji: "🐶", Color: "#123456"}
	_, err := d.CreateUser(req)
	if err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}
	_, err = d.CreateUser(req)
	if err == nil {
		t.Fatal("expected error on duplicate name, got nil")
	}
}

func TestGetAllUsers(t *testing.T) {
	d := setupDB(t)
	users, err := d.GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}

	_, _ = d.CreateUser(models.RegisterRequest{Name: "A", Emoji: "😊", Color: "#aaa"})
	_, _ = d.CreateUser(models.RegisterRequest{Name: "B", Emoji: "😊", Color: "#bbb"})

	users, err = d.GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	d := setupDB(t)
	u, err := d.CreateUser(models.RegisterRequest{Name: "Eve", Emoji: "🦊", Color: "#000"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	updated, err := d.UpdateUser(u.ID, models.UpdateProfileRequest{
		Name:     "Eve Updated",
		Emoji:    "🦁",
		Color:    "#fff",
		Theme:    "dark",
		Timezone: "Europe/Moscow",
		AutoTZ:   false,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.Name != "Eve Updated" {
		t.Errorf("got name %s, want Eve Updated", updated.Name)
	}
	if updated.Theme != "dark" {
		t.Errorf("got theme %s, want dark", updated.Theme)
	}
}

func TestDeleteUser(t *testing.T) {
	d := setupDB(t)
	u, _ := d.CreateUser(models.RegisterRequest{Name: "Del", Emoji: "😀", Color: "#111"})
	err := d.DeleteUser(u.ID)
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	_, err = d.GetUserByID(u.ID)
	if err == nil {
		t.Fatal("expected error getting deleted user, got nil")
	}
}

func TestSortOrder(t *testing.T) {
	d := setupDB(t)
	u1, _ := d.CreateUser(models.RegisterRequest{Name: "First", Emoji: "1️⃣", Color: "#111"})
	u2, _ := d.CreateUser(models.RegisterRequest{Name: "Second", Emoji: "2️⃣", Color: "#222"})
	u3, _ := d.CreateUser(models.RegisterRequest{Name: "Third", Emoji: "3️⃣", Color: "#333"})

	// Reverse order
	err := d.UpdateSortOrder([]int64{u3.ID, u1.ID, u2.ID})
	if err != nil {
		t.Fatalf("UpdateSortOrder: %v", err)
	}

	users, _ := d.GetAllUsers()
	if users[0].ID != u3.ID {
		t.Errorf("expected first user ID %d, got %d", u3.ID, users[0].ID)
	}
}

func TestSlotCRUD(t *testing.T) {
	d := setupDB(t)
	u, _ := d.CreateUser(models.RegisterRequest{Name: "SlotUser", Emoji: "⏰", Color: "#999"})

	now := time.Now().UTC().Truncate(time.Second)
	start := now.Add(time.Hour)
	end := now.Add(2 * time.Hour)

	slot, err := d.CreateSlot(u.ID, start, end, "test comment")
	if err != nil {
		t.Fatalf("CreateSlot: %v", err)
	}
	if slot.ID == 0 {
		t.Fatal("expected non-zero slot ID")
	}
	if slot.Comment != "test comment" {
		t.Errorf("got comment %q, want 'test comment'", slot.Comment)
	}

	// Update
	updated, err := d.UpdateSlot(slot.ID, start.Add(30*time.Minute), end, "updated")
	if err != nil {
		t.Fatalf("UpdateSlot: %v", err)
	}
	if updated.Comment != "updated" {
		t.Errorf("got comment %q, want 'updated'", updated.Comment)
	}

	// Get slots
	slots, err := d.GetSlots(u.ID, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetSlots: %v", err)
	}
	if len(slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(slots))
	}

	// Delete
	err = d.DeleteSlot(slot.ID)
	if err != nil {
		t.Fatalf("DeleteSlot: %v", err)
	}
	slots, _ = d.GetSlots(u.ID, time.Time{}, time.Time{})
	if len(slots) != 0 {
		t.Errorf("expected 0 slots after delete, got %d", len(slots))
	}
}

func TestEventCRUD(t *testing.T) {
	d := setupDB(t)
	u, _ := d.CreateUser(models.RegisterRequest{Name: "EventUser", Emoji: "🎉", Color: "#gold"})

	now := time.Now().UTC().Truncate(time.Second)
	start := now.Add(24 * time.Hour)
	end := now.Add(25 * time.Hour)

	event, err := d.CreateEvent(models.Event{
		Title:       "Team Meeting",
		Description: "Weekly sync",
		Icon:        "📅",
		Color:       "#D4AF37",
		StartTime:   start,
		EndTime:     end,
		CreatedBy:   u.ID,
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	if event.ID == 0 {
		t.Fatal("expected non-zero event ID")
	}
	if event.Title != "Team Meeting" {
		t.Errorf("got title %q, want 'Team Meeting'", event.Title)
	}

	// Update
	updated, err := d.UpdateEvent(event.ID, models.Event{
		Title: "Updated Meeting", Description: "Updated", Icon: "✅",
		Color: "#D4AF37", StartTime: start, EndTime: end,
	})
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	if updated.Title != "Updated Meeting" {
		t.Errorf("got title %q, want 'Updated Meeting'", updated.Title)
	}

	// Get events
	events, err := d.GetEvents(time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetEvents: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	// Delete
	err = d.DeleteEvent(event.ID)
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	events, _ = d.GetEvents(time.Time{}, time.Time{})
	if len(events) != 0 {
		t.Errorf("expected 0 events after delete, got %d", len(events))
	}
}

func TestCleanupOldData(t *testing.T) {
	d := setupDB(t)
	u, _ := d.CreateUser(models.RegisterRequest{Name: "CleanUser", Emoji: "🧹", Color: "#eee"})

	// Old slot (7 months ago)
	oldTime := time.Now().Add(-7 * 30 * 24 * time.Hour)
	_, _ = d.CreateSlot(u.ID, oldTime, oldTime.Add(time.Hour), "old")

	// Future slot
	futureTime := time.Now().Add(24 * time.Hour)
	_, _ = d.CreateSlot(u.ID, futureTime, futureTime.Add(time.Hour), "future")

	d.CleanupOldData()

	slots, _ := d.GetSlots(u.ID, time.Time{}, time.Time{})
	if len(slots) != 1 {
		t.Errorf("expected 1 slot after cleanup, got %d", len(slots))
	}
	if slots[0].Comment != "future" {
		t.Errorf("expected future slot to survive, got comment %q", slots[0].Comment)
	}
}
