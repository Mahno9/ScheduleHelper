package models

import "time"

// User represents a registered user
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Emoji     string    `json:"emoji"`
	Color     string    `json:"color"` // hex color, e.g. "#FF5733"
	Theme     string    `json:"theme"` // "light", "dark", "system"
	Timezone  string    `json:"timezone"`
	AutoTZ    bool      `json:"auto_timezone"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// Slot represents a free time slot for a user
type Slot struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event represents a shared calendar event
type Event struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"` // gold shade
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EventParticipant links users to events
type EventParticipant struct {
	EventID int64 `json:"event_id"`
	UserID  int64 `json:"user_id"`
}

// CalendarData is the aggregated view for the shared calendar
type CalendarData struct {
	Users  []User  `json:"users"`
	Slots  []Slot  `json:"slots"`
	Events []Event `json:"events"`
}

// LoginRequest for POST /api/login
type LoginRequest struct {
	UserID int64 `json:"user_id"`
}

// RegisterRequest for POST /api/register
type RegisterRequest struct {
	Name     string `json:"name"`
	Emoji    string `json:"emoji"`
	Color    string `json:"color"`
	Timezone string `json:"timezone"`
	AutoTZ   bool   `json:"auto_timezone"`
}

// UpdateProfileRequest for PUT /api/profile/:id
type UpdateProfileRequest struct {
	Name      string `json:"name"`
	Emoji     string `json:"emoji"`
	Color     string `json:"color"`
	Theme     string `json:"theme"`
	Timezone  string `json:"timezone"`
	AutoTZ    bool   `json:"auto_timezone"`
}

// UpdateSortOrderRequest for PUT /api/users/sort
type UpdateSortOrderRequest struct {
	Order []int64 `json:"order"` // user IDs in new sort order
}

// SSEMessage for server-sent events
type SSEMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
