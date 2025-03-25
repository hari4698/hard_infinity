package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hari4698/hardinfinity/internal/auth"
	"github.com/hari4698/hardinfinity/internal/db"
	"github.com/hari4698/hardinfinity/internal/models"
	"github.com/hari4698/hardinfinity/internal/utils"
)

// GetChallenges retrieves all challenges for the authenticated user
func GetChallenges(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var userID uuid.UUID
	err := db.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	rows, err := db.DB.Query(r.Context(), `
		SELECT id, name, description, start_date, end_date, current_day, status, created_at, updated_at
				FROM challenges
				WHERE user_id = $1
				ORDER BY created_at DESC
		`, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve challenges")
		return
	}

	defer rows.Close()

	challenges := []models.Challenge{}
	for rows.Next() {
		var c models.Challenge
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.StartDate, &c.EndDate, &c.CurrentDay, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to scan challenge data")
			return
		}

		c.UserID = userID
		challenges = append(challenges, c)
	}

	utils.Success(w, http.StatusOK, challenges)
}

func GetChallenge(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var userID uuid.UUID
	err := db.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	challengeID := chi.URLParam(r, "id")
	if challengeID == "" {
		utils.Error(w, http.StatusBadGateway, "Challenge ID is required")
		return
	}

	challengeUUID, err := uuid.Parse(challengeID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID format")
		return
	}

	var challenge models.Challenge
	err = db.DB.QueryRow(r.Context(), `
		SELECT id, name, description, start_date, end_date, current_day, status, created_at, updated_at
				FROM challenges
				WHERE id = $1 AND user_id = $2
		`, challengeUUID, userID).Scan(&challenge.ID, &challenge.Name, &challenge.Description, &challenge.StartDate, &challenge.EndDate, &challenge.CurrentDay, &challenge.Status, &challenge.CreatedAt, &challenge.UpdatedAt)

	if err != nil {
		utils.Error(w, http.StatusNotFound, "Chanllenge not found")
		return
	}

	challenge.UserID = userID
	utils.Success(w, http.StatusOK, challenge)
}

// CreateChallenge creates a new challenge
func CreateChallenge(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query to get user ID from clerk_id
	var userID uuid.UUID
	err := db.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	var challenge models.Challenge
	if err := json.NewDecoder(r.Body).Decode(&challenge); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	challenge.ID = uuid.New()
	challenge.UserID = userID
	challenge.Status = "active"
	challenge.CurrentDay = 1
	challenge.CreatedAt = time.Now()
	challenge.UpdatedAt = time.Now()

	_, err = db.DB.Exec(r.Context(), `
			INSERT INTO challenges (id, user_id, name, description, start_date, end_date, current_day, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, challenge.ID, challenge.UserID, challenge.Name, challenge.Description, challenge.StartDate, challenge.EndDate, challenge.CurrentDay, challenge.Status, challenge.CreatedAt, challenge.UpdatedAt)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create challenge")
		return
	}

	utils.Success(w, http.StatusCreated, challenge)
}

// UpdateChallenge updates an existing challenge
func UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query to get user ID from clerk_id
	var userID uuid.UUID
	err := db.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	challengeID := chi.URLParam(r, "id")
	if challengeID == "" {
		utils.Error(w, http.StatusBadRequest, "Challenge ID is required")
		return
	}

	challengeUUID, err := uuid.Parse(challengeID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID format")
		return
	}

	// Check if the challenge exists and belongs to the user
	var exists bool
	err = db.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM challenges WHERE id = $1 AND user_id = $2)", challengeUUID, userID).Scan(&exists)
	if err != nil || !exists {
		utils.Error(w, http.StatusNotFound, "Challenge not found")
		return
	}

	var challenge models.Challenge
	if err := json.NewDecoder(r.Body).Decode(&challenge); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	challenge.ID = challengeUUID
	challenge.UserID = userID
	challenge.UpdatedAt = time.Now()

	_, err = db.DB.Exec(r.Context(), `
		UPDATE challenges 
		SET name = $1, description = $2, start_date = $3, end_date = $4, current_day = $5, status = $6, updated_at = $7
		WHERE id = $8 AND user_id = $9
	`, challenge.Name, challenge.Description, challenge.StartDate, challenge.EndDate, challenge.CurrentDay, challenge.Status, challenge.UpdatedAt, challenge.ID, challenge.UserID)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update challenge")
		return
	}

	utils.Success(w, http.StatusOK, challenge)
}

// DeleteChallenge deletes a challenge by ID
func DeleteChallenge(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query to get user ID from clerk_id
	var userID uuid.UUID
	err := db.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	challengeID := chi.URLParam(r, "id")
	if challengeID == "" {
		utils.Error(w, http.StatusBadRequest, "Challenge ID is required")
		return
	}

	challengeUUID, err := uuid.Parse(challengeID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID format")
		return
	}

	// Start a transaction for cascading deletes
	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback(context.Background())

	// Delete all related records (this would be better handled with SQL CASCADE)
	// First, get all sections
	rows, err := tx.Query(r.Context(), "SELECT id FROM sections WHERE challenge_id = $1", challengeUUID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query sections")
		return
	}

	var sectionIDs []uuid.UUID
	for rows.Next() {
		var sectionID uuid.UUID
		if err := rows.Scan(&sectionID); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to scan section ID")
			return
		}
		sectionIDs = append(sectionIDs, sectionID)
	}
	rows.Close()

	// Delete tasks for each section
	for _, sectionID := range sectionIDs {
		_, err = tx.Exec(r.Context(), "DELETE FROM tasks WHERE section_id = $1", sectionID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to delete tasks")
			return
		}
	}

	// Get all daily entries
	rows, err = tx.Query(r.Context(), "SELECT id FROM daily_entries WHERE challenge_id = $1", challengeUUID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query daily entries")
		return
	}

	var entryIDs []uuid.UUID
	for rows.Next() {
		var entryID uuid.UUID
		if err := rows.Scan(&entryID); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to scan entry ID")
			return
		}
		entryIDs = append(entryIDs, entryID)
	}
	rows.Close()

	// Delete task entries for each daily entry
	for _, entryID := range entryIDs {
		_, err = tx.Exec(r.Context(), "DELETE FROM task_entries WHERE daily_entry_id = $1", entryID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to delete task entries")
			return
		}
	}

	// Delete daily entries
	_, err = tx.Exec(r.Context(), "DELETE FROM daily_entries WHERE challenge_id = $1", challengeUUID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete daily entries")
		return
	}

	// Delete measurements
	_, err = tx.Exec(r.Context(), "DELETE FROM measurements WHERE challenge_id = $1", challengeUUID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete measurements")
		return
	}

	// Delete sections
	_, err = tx.Exec(r.Context(), "DELETE FROM sections WHERE challenge_id = $1", challengeUUID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete sections")
		return
	}

	// Finally delete the challenge
	result, err := tx.Exec(r.Context(), "DELETE FROM challenges WHERE id = $1 AND user_id = $2", challengeUUID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete challenge")
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		utils.Error(w, http.StatusNotFound, "Challenge not found")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]string{"message": "Challenge deleted successfully"})
}

// ResetChallenge resets a challenge to day 1
func ResetChallenge(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query to get user ID from clerk_id
	var userID uuid.UUID
	err := db.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	challengeID := chi.URLParam(r, "id")
	if challengeID == "" {
		utils.Error(w, http.StatusBadRequest, "Challenge ID is required")
		return
	}

	challengeUUID, err := uuid.Parse(challengeID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID format")
		return
	}

	// Start a transaction
	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback(context.Background())

	// Reset challenge to day 1
	_, err = tx.Exec(r.Context(), `
		UPDATE challenges 
		SET current_day = 1, status = 'active', updated_at = $1
		WHERE id = $2 AND user_id = $3
	`, time.Now(), challengeUUID, userID)
	
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to reset challenge")
		return
	}

	// Delete all daily entries and task entries
	_, err = tx.Exec(r.Context(), `
		DELETE FROM task_entries 
		WHERE daily_entry_id IN (
			SELECT id FROM daily_entries WHERE challenge_id = $1
		)
	`, challengeUUID)
	
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to clear task entries")
		return
	}

	_, err = tx.Exec(r.Context(), "DELETE FROM daily_entries WHERE challenge_id = $1", challengeUUID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to clear daily entries")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]string{"message": "Challenge reset successfully"})
}

// GetChallengeProgress retrieves progress statistics for a challenge
func GetChallengeProgress(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query to get user ID from clerk_id
	var userID uuid.UUID
	err := db.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	challengeID := chi.URLParam(r, "id")
	if challengeID == "" {
		utils.Error(w, http.StatusBadRequest, "Challenge ID is required")
		return
	}

	challengeUUID, err := uuid.Parse(challengeID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid challenge ID format")
		return
	}

	// First, check if challenge exists and belongs to user
	var challenge models.Challenge
	err = db.DB.QueryRow(r.Context(), `
		SELECT id, name, current_day, status 
		FROM challenges 
		WHERE id = $1 AND user_id = $2
	`, challengeUUID, userID).Scan(&challenge.ID, &challenge.Name, &challenge.CurrentDay, &challenge.Status)
	
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Challenge not found")
		return
	}

	// Get completed days count
	var completedDays int
	err = db.DB.QueryRow(r.Context(), `
		SELECT COUNT(*) 
		FROM daily_entries 
		WHERE challenge_id = $1 AND completed = true
	`, challengeUUID).Scan(&completedDays)
	
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve progress data")
		return
	}

	// Get streak information
	var currentStreak, longestStreak int
	rows, err := db.DB.Query(r.Context(), `
		SELECT day_number, completed, date
		FROM daily_entries
		WHERE challenge_id = $1
		ORDER BY day_number ASC
	`, challengeUUID)
	
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve streak data")
		return
	}
	defer rows.Close()

	var entries []struct {
		DayNumber int       `json:"day_number"`
		Completed bool      `json:"completed"`
		Date      time.Time `json:"date"`
	}

	for rows.Next() {
		var entry struct {
			DayNumber int       `json:"day_number"`
			Completed bool      `json:"completed"`
			Date      time.Time `json:"date"`
		}
		if err := rows.Scan(&entry.DayNumber, &entry.Completed, &entry.Date); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to scan entry data")
			return
		}
		entries = append(entries, entry)
	}

	// Calculate streaks
	streak := 0
	for i, entry := range entries {
		if entry.Completed {
			streak++
			if streak > longestStreak {
				longestStreak = streak
			}
		} else {
			streak = 0
		}

		// If this is the last entry or there's a day gap, check if it's the current streak
		if i == len(entries)-1 || entries[i+1].DayNumber > entry.DayNumber+1 {
			if entry.Completed && streak > 0 {
				currentStreak = streak
			}
		}
	}

	// Create progress response
	progress := struct {
		TotalDays      int    `json:"total_days"`
		CurrentDay     int    `json:"current_day"`
		CompletedDays  int    `json:"completed_days"`
		CurrentStreak  int    `json:"current_streak"`
		LongestStreak  int    `json:"longest_streak"`
		Status         string `json:"status"`
		CompletionRate float64 `json:"completion_rate"`
	}{
		TotalDays:      75, // Default for Hard75, could be customized
		CurrentDay:     challenge.CurrentDay,
		CompletedDays:  completedDays,
		CurrentStreak:  currentStreak,
		LongestStreak:  longestStreak,
		Status:         challenge.Status,
		CompletionRate: float64(completedDays) / 75.0 * 100, // Calculate completion percentage
	}

	utils.Success(w, http.StatusOK, progress)
}

