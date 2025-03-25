package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hari4698/hardinfinity/internal/db"
	"github.com/hari4698/hardinfinity/internal/models"
	"github.com/hari4698/hardinfinity/internal/utils"
)

// GetMeasurements retrieves all measurements for a specific challenge
func GetMeasurements(w http.ResponseWriter, r *http.Request) {
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
		`SELECT id, challenge_id, day_number, date, weight, chest, waist,
		hips, arms, thighs, created_at, updated_at
		FROM measurements
		WHERE challenge_id = $1
		ORDER BY date ASC`,
		challengeID)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve measurements")
		return
	}
	defer rows.Close()

	measurements := []models.Measurement{}
	for rows.Next() {
		var measurement models.Measurement
		if err := rows.Scan(
			&measurement.ID,
			&measurement.ChallengeID,
			&measurement.DayNumber,
			&measurement.Date,
			&measurement.Weight,
			&measurement.Chest,
			&measurement.Waist,
			&measurement.Hips,
			&measurement.Arms,
			&measurement.Thighs,
			&measurement.CreatedAt,
			&measurement.UpdatedAt,
		); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Error parsing measurement data")
			return
		}
		measurements = append(measurements, measurement)
	}

	utils.Success(w, http.StatusOK, measurements)
}

// AddMeasurement adds a new measurement for a challenge
func AddMeasurement(w http.ResponseWriter, r *http.Request) {
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
		`SELECT id, user_id, current_day, status
		FROM challenges WHERE id = $1`,
		challengeID).Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.CurrentDay,
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

	// Parse request body
	var measurement models.Measurement
	if err := json.NewDecoder(r.Body).Decode(&measurement); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set challenge ID and generate a new UUID
	measurement.ID = uuid.New()
	measurement.ChallengeID = challengeID

	// If day number is not provided, use the current day of the challenge
	if measurement.DayNumber <= 0 {
		measurement.DayNumber = challenge.CurrentDay
	}

	// If date is not provided, use today's date
	if measurement.Date.IsZero() {
		measurement.Date = time.Now()
	}

	// Current time for created_at and updated_at
	now := time.Now()
	measurement.CreatedAt = now
	measurement.UpdatedAt = now

	// Insert the new measurement
	_, err = db.DB.Exec(ctx,
		`INSERT INTO measurements
		(id, challenge_id, day_number, date, weight, chest, waist,
		hips, arms, thighs, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		measurement.ID,
		measurement.ChallengeID,
		measurement.DayNumber,
		measurement.Date,
		measurement.Weight,
		measurement.Chest,
		measurement.Waist,
		measurement.Hips,
		measurement.Arms,
		measurement.Thighs,
		measurement.CreatedAt,
		measurement.UpdatedAt)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to save measurement")
		return
	}

	utils.Success(w, http.StatusCreated, measurement)
}

// UpdateMeasurement updates an existing measurement
func UpdateMeasurement(w http.ResponseWriter, r *http.Request) {
	measurementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid measurement ID")
		return
	}

	ctx := r.Context()

	// First verify the measurement exists and get its challenge ID
	var challengeID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT challenge_id FROM measurements WHERE id = $1",
		measurementID).Scan(&challengeID)

	if err != nil {
		utils.Error(w, http.StatusNotFound, "Measurement not found")
		return
	}

	// Verify user has access to this challenge
	userID := r.Context().Value("userID").(uuid.UUID)
	var ownerID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT user_id FROM challenges WHERE id = $1",
		challengeID).Scan(&ownerID)

	if err != nil || ownerID != userID {
		utils.Error(w, http.StatusForbidden, "You don't have access to this measurement")
		return
	}

	// Parse request body
	var measurementUpdate struct {
		DayNumber int       `json:"day_number"`
		Date      time.Time `json:"date"`
		Weight    float64   `json:"weight"`
		Chest     float64   `json:"chest"`
		Waist     float64   `json:"waist"`
		Hips      float64   `json:"hips"`
		Arms      float64   `json:"arms"`
		Thighs    float64   `json:"thighs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&measurementUpdate); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update the measurement
	_, err = db.DB.Exec(ctx,
		`UPDATE measurements 
		SET day_number = $1, date = $2, weight = $3, chest = $4, waist = $5, 
		hips = $6, arms = $7, thighs = $8, updated_at = NOW()
		WHERE id = $9`,
		measurementUpdate.DayNumber,
		measurementUpdate.Date,
		measurementUpdate.Weight,
		measurementUpdate.Chest,
		measurementUpdate.Waist,
		measurementUpdate.Hips,
		measurementUpdate.Arms,
		measurementUpdate.Thighs,
		measurementID)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update measurement")
		return
	}

	// Retrieve the updated measurement
	var updatedMeasurement models.Measurement
	err = db.DB.QueryRow(ctx,
		`SELECT id, challenge_id, day_number, date, weight, chest, waist, 
		hips, arms, thighs, created_at, updated_at
		FROM measurements 
		WHERE id = $1`,
		measurementID).Scan(
		&updatedMeasurement.ID,
		&updatedMeasurement.ChallengeID,
		&updatedMeasurement.DayNumber,
		&updatedMeasurement.Date,
		&updatedMeasurement.Weight,
		&updatedMeasurement.Chest,
		&updatedMeasurement.Waist,
		&updatedMeasurement.Hips,
		&updatedMeasurement.Arms,
		&updatedMeasurement.Thighs,
		&updatedMeasurement.CreatedAt,
		&updatedMeasurement.UpdatedAt,
	)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve updated measurement")
		return
	}

	utils.Success(w, http.StatusOK, updatedMeasurement)
}

// DeleteMeasurement deletes a measurement
func DeleteMeasurement(w http.ResponseWriter, r *http.Request) {
	measurementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid measurement ID")
		return
	}

	ctx := r.Context()

	// First verify the measurement exists and get its challenge ID
	var challengeID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT challenge_id FROM measurements WHERE id = $1",
		measurementID).Scan(&challengeID)

	if err != nil {
		utils.Error(w, http.StatusNotFound, "Measurement not found")
		return
	}

	// Verify user has access to this challenge
	userID := r.Context().Value("userID").(uuid.UUID)
	var ownerID uuid.UUID
	err = db.DB.QueryRow(ctx,
		"SELECT user_id FROM challenges WHERE id = $1",
		challengeID).Scan(&ownerID)

	if err != nil || ownerID != userID {
		utils.Error(w, http.StatusForbidden, "You don't have access to this measurement")
		return
	}

	// Delete the measurement
	_, err = db.DB.Exec(ctx,
		"DELETE FROM measurements WHERE id = $1",
		measurementID)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete measurement")
		return
	}

	utils.Success(w, http.StatusOK, map[string]any{
		"message": "Measurement deleted successfully",
		"id":      measurementID,
	})
}