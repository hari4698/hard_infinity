package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hari4698/hardinfinity/internal/db"
	"github.com/hari4698/hardinfinity/internal/models"
	"github.com/hari4698/hardinfinity/internal/utils"
)

func GetDailyEntries(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	ctx := r.Context()

	// Verify user has access to this challenge
	userID := r.Context().Value("userID").(uuid.UUID)
	var ownerID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT user_id FROM challenges WHERE id = $1",
		challengeID).Scan(&ownerID)

	if err != nil || ownerID != userID {
		utils.Error(w, http.StatusForbidden, "You don't have access to this challenge")
		return
	}

	rows, err := db.DB.Query(ctx,
		`SELECT id, challenge_id, day_number, date, completed, notes,
			progress_photo_url, energy_level, mood_level, created_at, updated_at
			FROM daily_entries
			WHERE challenge_id = $1
			ORDER BY day_number ASC`,
		challengeID)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve daily entries")
		return
	}
	defer rows.Close()

	entries := []models.DailyEntry{}
	for rows.Next() {
		var entry models.DailyEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.ChallengeID,
			&entry.DayNumber,
			&entry.Date,
			&entry.Completed,
			&entry.Notes,
			&entry.ProgressPhotoURL,
			&entry.EnergyLevel,
			&entry.MoodLevel,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Error parsing daily entry data")
			return
		}
		entries = append(entries, entry)
	}

	utils.Success(w, http.StatusOK, entries)
}

// GetDailyEntry retrieves a specific daily entry by day number
func GetDailyEntry(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	dayNumber, err := strconv.Atoi(chi.URLParam(r, "day"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid day number")
		return
	}

	ctx := r.Context()

	// Verify user has access to this challenge
	userID := r.Context().Value("userID").(uuid.UUID)
	var ownerID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT user_id FROM challenges WHERE id = $1",
		challengeID).Scan(&ownerID)

	if err != nil || ownerID != userID {
		utils.Error(w, http.StatusForbidden, "You don't have access to this challenge")
		return
	}

	var entry models.DailyEntry
	err = db.DB.QueryRow(ctx,
		`SELECT id, challenge_id, day_number, date, completed, notes,
		progress_photo_url, energy_level, mood_level, created_at, updated_at
		FROM daily_entries
		WHERE challenge_id = $1 AND day_number = $2`,
		challengeID, dayNumber).Scan(
		&entry.ID,
		&entry.ChallengeID,
		&entry.DayNumber,
		&entry.Date,
		&entry.Completed,
		&entry.Notes,
		&entry.ProgressPhotoURL,
		&entry.EnergyLevel,
		&entry.MoodLevel,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		utils.Error(w, http.StatusNotFound, "Entry not found")
		return
	}

	// Get all task entries for this daily entry
	rows, err := db.DB.Query(ctx,
		`SELECT id, daily_entry_id, task_id, completed, value, notes, created_at, updated_at
		FROM task_entries
		WHERE daily_entry_id = $1`,
		entry.ID)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve task entries")
		return
	}
	defer rows.Close()

	taskEntries := []models.TaskEntry{}
	for rows.Next() {
		var taskEntry models.TaskEntry
		var valueJSON []byte

		if err := rows.Scan(
			&taskEntry.ID,
			&taskEntry.DailyEntryID,
			&taskEntry.TaskID,
			&taskEntry.Completed,
			&valueJSON,
			&taskEntry.Notes,
			&taskEntry.CreatedAt,
			&taskEntry.UpdatedAt,
		); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Error parsing task entry data")
			return
		}

		// Parse the JSON value field
		if len(valueJSON) > 0 {
			json.Unmarshal(valueJSON, &taskEntry.Value)
		}

		taskEntries = append(taskEntries, taskEntry)
	}

	// Combine daily entry with task entries
	result := map[string]any{
		"entry":       entry,
		"taskEntries": taskEntries,
	}

	utils.Success(w, http.StatusOK, result)
}

// CreateOrUpdateTodayEntry creates or updates an entry for today
func CreateOrUpdateTodayEntry(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	ctx := r.Context()

	// Verify user has access to this challenge
	userID := r.Context().Value("userID").(uuid.UUID)
	var challenge models.Challenge
	err = db.DB.QueryRow(ctx,
		`SELECT id, user_id, current_day, start_date, status
		FROM challenges WHERE id = $1`,
		challengeID).Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.CurrentDay,
		&challenge.StartDate,
		&challenge.Status,
	)

	if err != nil {
		utils.Error(w, http.StatusNotFound, "Challenge not found")
		return
	}

	if challenge.UserID != userID {
		utils.Error(w, http.StatusForbidden, "You don't have access to this challenge")
		return
	}

	if challenge.Status != "active" {
		utils.Error(w, http.StatusBadRequest, "Can only add entries to active challenges")
		return
	}

	// Parse request body
	var entryData struct {
		Completed        bool                     `json:"completed"`
		Notes            string                   `json:"notes"`
		ProgressPhotoURL string                   `json:"progress_photo_url"`
		EnergyLevel      int                      `json:"energy_level"`
		MoodLevel        int                      `json:"mood_level"`
		TaskEntries      []map[string]any `json:"task_entries"`
	}

	if err := json.NewDecoder(r.Body).Decode(&entryData); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Current day for the challenge
	dayNumber := challenge.CurrentDay
	today := time.Now()

	// Start a transaction
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	// Check if entry exists for today
	var entryID uuid.UUID
	var exists bool
	err = tx.QueryRow(ctx,
		"SELECT id FROM daily_entries WHERE challenge_id = $1 AND day_number = $2",
		challengeID, dayNumber).Scan(&entryID)

	if err == nil {
		// Entry exists, update it
		exists = true
		_, err = tx.Exec(ctx,
			`UPDATE daily_entries
			SET completed = $1, notes = $2, progress_photo_url = $3,
			energy_level = $4, mood_level = $5, updated_at = NOW()
			WHERE id = $6`,
			entryData.Completed, entryData.Notes, entryData.ProgressPhotoURL,
			entryData.EnergyLevel, entryData.MoodLevel, entryID)
	} else {
		// Create new entry
		exists = false
		entryID = uuid.New()
		_, err = tx.Exec(ctx,
			`INSERT INTO daily_entries
			(id, challenge_id, day_number, date, completed, notes,
			progress_photo_url, energy_level, mood_level, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())`,
			entryID, challengeID, dayNumber, today, entryData.Completed, entryData.Notes,
			entryData.ProgressPhotoURL, entryData.EnergyLevel, entryData.MoodLevel)
	}

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to save daily entry")
		return
	}

	// Process task entries
	for _, taskEntryData := range entryData.TaskEntries {
		taskID, err := uuid.Parse(taskEntryData["task_id"].(string))
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid task ID in task entries")
			return
		}

		completed, _ := taskEntryData["completed"].(bool)
		notes, _ := taskEntryData["notes"].(string)
		value := taskEntryData["value"]

		// Convert value to JSON
		valueJSON, err := json.Marshal(value)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to process task entry value")
			return
		}

		// Check if task entry exists
		var taskEntryID uuid.UUID
		err = tx.QueryRow(ctx,
			"SELECT id FROM task_entries WHERE daily_entry_id = $1 AND task_id = $2",
			entryID, taskID).Scan(&taskEntryID)

		if err == nil {
			// Update existing task entry
			_, err = tx.Exec(ctx,
				`UPDATE task_entries
				SET completed = $1, value = $2, notes = $3, updated_at = NOW()
				WHERE id = $4`,
				completed, valueJSON, notes, taskEntryID)
		} else {
			// Create new task entry
			taskEntryID = uuid.New()
			_, err = tx.Exec(ctx,
				`INSERT INTO task_entries
				(id, daily_entry_id, task_id, completed, value, notes, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
				taskEntryID, entryID, taskID, completed, valueJSON, notes)
		}

		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to save task entry")
			return
		}
	}

	// If entry is marked as completed and it's the first time, increment current_day in challenge
	if entryData.Completed && !exists {
		_, err = tx.Exec(ctx,
			"UPDATE challenges SET current_day = current_day + 1, updated_at = NOW() WHERE id = $1",
			challengeID)

		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to update challenge progress")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]any{
		"entry_id":   entryID,
		"day_number": dayNumber,
		"message":    "Daily entry saved successfully",
	})
}

// UpdateDailyEntry updates a specific daily entry by day number
func UpdateDailyEntry(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID")
		return
	}

	dayNumber, err := strconv.Atoi(chi.URLParam(r, "day"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid day number")
		return
	}

	ctx := r.Context()

	// Verify user has access to this challenge
	userID := r.Context().Value("userID").(uuid.UUID)
	var ownerID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT user_id FROM challenges WHERE id = $1",
		challengeID).Scan(&ownerID)

	if err != nil || ownerID != userID {
		utils.Error(w, http.StatusForbidden, "You don't have access to this challenge")
		return
	}

	// Parse request body
	var entryData struct {
		Completed        bool                     `json:"completed"`
		Notes            string                   `json:"notes"`
		ProgressPhotoURL string                   `json:"progress_photo_url"`
		EnergyLevel      int                      `json:"energy_level"`
		MoodLevel        int                      `json:"mood_level"`
		TaskEntries      []map[string]any `json:"task_entries"`
	}

	if err := json.NewDecoder(r.Body).Decode(&entryData); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if entry exists
	var entryID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT id FROM daily_entries WHERE challenge_id = $1 AND day_number = $2",
		challengeID, dayNumber).Scan(&entryID)

	if err != nil {
		utils.Error(w, http.StatusNotFound, "Entry not found")
		return
	}

	// Start a transaction
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	// Update the daily entry
	_, err = tx.Exec(ctx,
		`UPDATE daily_entries
		SET completed = $1, notes = $2, progress_photo_url = $3,
		energy_level = $4, mood_level = $5, updated_at = NOW()
		WHERE id = $6`,
		entryData.Completed, entryData.Notes, entryData.ProgressPhotoURL,
		entryData.EnergyLevel, entryData.MoodLevel, entryID)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update daily entry")
		return
	}

	// Process task entries
	for _, taskEntryData := range entryData.TaskEntries {
		taskID, err := uuid.Parse(taskEntryData["task_id"].(string))
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid task ID in task entries")
			return
		}

		completed, _ := taskEntryData["completed"].(bool)
		notes, _ := taskEntryData["notes"].(string)
		value := taskEntryData["value"]

		// Convert value to JSON
		valueJSON, err := json.Marshal(value)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to process task entry value")
			return
		}

		// Check if task entry exists
		var taskEntryID uuid.UUID
		err = tx.QueryRow(ctx,
			"SELECT id FROM task_entries WHERE daily_entry_id = $1 AND task_id = $2",
			entryID, taskID).Scan(&taskEntryID)

		if err == nil {
			// Update existing task entry
			_, err = tx.Exec(ctx,
				`UPDATE task_entries
				SET completed = $1, value = $2, notes = $3, updated_at = NOW()
				WHERE id = $4`,
				completed, valueJSON, notes, taskEntryID)
		} else {
			// Create new task entry
			taskEntryID = uuid.New()
			_, err = tx.Exec(ctx,
				`INSERT INTO task_entries
				(id, daily_entry_id, task_id, completed, value, notes, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
				taskEntryID, entryID, taskID, completed, valueJSON, notes)
		}

		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to save task entry")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]any{
		"entry_id":   entryID,
		"day_number": dayNumber,
		"message":    "Daily entry updated successfully",
	})
}
