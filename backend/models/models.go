package models

import "time"

type User struct {
	ID                  string    `json:"id"`
	Username            string    `json:"username"`
	Color               string    `json:"color"`
	Emoji               string    `json:"emoji"`
	Theme               string    `json:"theme"`
	Timezone            string    `json:"timezone"`
	PasswordHash        *string   `json:"-"`
	GoogleCalendarToken *string   `json:"-"`
	CreatedAt           time.Time `json:"created_at"`
}

type Slot struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Comment   string    `json:"comment"`
}

type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

type EventParticipant struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
}