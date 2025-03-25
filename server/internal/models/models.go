package models

import (
	"time"
	
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
		ClerkID   string    `json:"clerk_id"`
		Email     string    `json:"email"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
}

type Challenge struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CurrentDay  int       `json:"current_day"`
	Status      string    `json:"status"` // active, completed, failed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Section struct {
	ID          uuid.UUID `json:"id"`
	ChallengeID uuid.UUID `json:"challenge_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Task struct {
	ID            uuid.UUID `json:"id"`
	SectionID     uuid.UUID `json:"section_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	TaskType      string    `json:"task_type"` // boolean, number, text, select, etc.
	Required      bool      `json:"required"`
	RestartOnFail bool      `json:"restart_on_fail"`
	StrikesEnabled bool     `json:"strikes_enabled"`
	StrikesLimit   int      `json:"strikes_limit"`
	Order         int       `json:"order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DailyEntry struct {
	ID              uuid.UUID `json:"id"`
	ChallengeID     uuid.UUID `json:"challenge_id"`
	DayNumber       int       `json:"day_number"`
	Date            time.Time `json:"date"`
	Completed       bool      `json:"completed"`
	Notes           string    `json:"notes"`
	ProgressPhotoURL string   `json:"progress_photo_url"`
	EnergyLevel     int       `json:"energy_level"`
	MoodLevel       int       `json:"mood_level"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TaskEntry struct {
	ID           uuid.UUID    `json:"id"`
	DailyEntryID uuid.UUID    `json:"daily_entry_id"`
	TaskID       uuid.UUID    `json:"task_id"`
	Completed    bool         `json:"completed"`
	Value        interface{}  `json:"value"` // This will be handled as JSON
	Notes        string       `json:"notes"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Measurement struct {
	ID          uuid.UUID `json:"id"`
	ChallengeID uuid.UUID `json:"challenge_id"`
	DayNumber   int       `json:"day_number"`
	Date        time.Time `json:"date"`
	Weight      float64   `json:"weight"`
	Chest       float64   `json:"chest"`
	Waist       float64   `json:"waist"`
	Hips        float64   `json:"hips"`
	Arms        float64   `json:"arms"`
	Thighs      float64   `json:"thighs"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}