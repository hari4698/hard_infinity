package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hari4698/hardinfinity/internal/auth"
	"github.com/hari4698/hardinfinity/internal/db"
	"github.com/hari4698/hardinfinity/internal/models"
	"github.com/hari4698/hardinfinity/internal/utils"
)

type CreateSectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       string `json:"order"`
}

type UpdateSectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ReorderSectionRequest struct {
	Order int `json:"order"`
}

// GetSections retrieves all sections for a challenge
func GetSections(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "id")

	// Verify that the challenge exists and belongs to the authenticated user
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check challenge ownership
	if err := validateChallengeOwnership(challengeID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Challenge not found")
		return
	}

	// Query the database for sections
	rows, err := db.DB.Query(r.Context(), `
			SELECT id, challenge_id, name, description, "order", created_at, updated_at
			FROM sections
			WHERE challenge_id = $1
			ORDER BY "order" ASC
		`, challengeID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve sections")
		return
	}
	defer rows.Close()

	sections := []models.Section{}
	for rows.Next() {
		var section models.Section
		if err := rows.Scan(
			&section.ID,
			&section.ChallengeID,
			&section.Name,
			&section.Description,
			&section.Order,
			&section.CreatedAt,
			&section.UpdatedAt,
		); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Error scanning section data")
			return
		}
		sections = append(sections, section)
	}

	if err := rows.Err(); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Error iterating through sections")
		return
	}

	utils.Success(w, http.StatusOK, sections)
}

// CreateSection creates a new section for a challenge
func CreateSection(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "id")

	// Verify that the challenge exists and belongs to the authenticated user
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check challenge ownership
	if err := validateChallengeOwnership(challengeID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Challenge not found")
		return
	}

	// Parse request body
	var req CreateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		utils.Error(w, http.StatusBadRequest, "Section name is required")
		return
	}

	// Generate a new UUID for the section
	sectionID := uuid.New().String()

	// If order is not specified, determine the next available order
	if req.Order == "" || req.Order == "0" {
		var maxOrder int
		err := db.DB.QueryRow(r.Context(), `
			SELECT COALESCE(MAX("order"), 0) FROM sections WHERE challenge_id = $1
		`, challengeID).Scan(&maxOrder)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to determine section order")
			return
		}
		req.Order = string(maxOrder + 1)
	}

	// Insert the new section
	var section models.Section
	err := db.DB.QueryRow(r.Context(), `
		INSERT INTO sections (id, challenge_id, name, description, "order", created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, challenge_id, name, description, "order", created_at, updated_at
	`, sectionID, challengeID, req.Name, req.Description, req.Order).Scan(
		&section.ID,
		&section.ChallengeID,
		&section.Name,
		&section.Description,
		&section.Order,
		&section.CreatedAt,
		&section.UpdatedAt,
	)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create section")
		return
	}

	utils.Success(w, http.StatusCreated, section)
}

// UpdateSection updates an existing section
func UpdateSection(w http.ResponseWriter, r *http.Request) {
	sectionID := chi.URLParam(r, "id")

	// Verify user is authenticated
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check section ownership through challenge
	if err := validateSectionOwnership(sectionID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Section not found")
		return
	}

	// Parse request body
	var req UpdateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		utils.Error(w, http.StatusBadRequest, "Section name is required")
		return
	}

	// Update the section
	var section models.Section
	err := db.DB.QueryRow(r.Context(), `
		UPDATE sections
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, challenge_id, name, description, "order", created_at, updated_at
	`, req.Name, req.Description, sectionID).Scan(
		&section.ID,
		&section.ChallengeID,
		&section.Name,
		&section.Description,
		&section.Order,
		&section.CreatedAt,
		&section.UpdatedAt,
	)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update section")
		return
	}

	utils.Success(w, http.StatusOK, section)
}

// DeleteSection deletes a section and all its tasks
func DeleteSection(w http.ResponseWriter, r *http.Request) {
	sectionID := chi.URLParam(r, "id")

	// Verify user is authenticated
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check section ownership through challenge
	if err := validateSectionOwnership(sectionID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Section not found")
		return
	}

	// Start a transaction to delete tasks and section
	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	// Delete tasks associated with the section
	_, err = tx.Exec(r.Context(), "DELETE FROM tasks WHERE section_id = $1", sectionID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete tasks")
		return
	}

	// Delete the section
	_, err = tx.Exec(r.Context(), "DELETE FROM sections WHERE id = $1", sectionID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete section")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]string{"message": "Section deleted successfully"})
}

// ReorderSection changes the order of a section
func ReorderSection(w http.ResponseWriter, r *http.Request) {
	sectionID := chi.URLParam(r, "id")

	// Verify user is authenticated
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check section ownership through challenge
	if err := validateSectionOwnership(sectionID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Section not found")
		return
	}

	// Parse request body
	var req ReorderSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Order <= 0 {
		utils.Error(w, http.StatusBadRequest, "Order must be a positive integer")
		return
	}

	// Get the section's challenge ID
	var challengeID string
	err := db.DB.QueryRow(r.Context(), "SELECT challenge_id FROM sections WHERE id = $1", sectionID).Scan(&challengeID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get section info")
		return
	}

	// Update the section's order
	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	// Get current order
	var currentOrder int
	err = tx.QueryRow(r.Context(), `SELECT "order" FROM sections WHERE id = $1`, sectionID).Scan(&currentOrder)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get current order")
		return
	}

	// If moving up (smaller order number)
	if req.Order < currentOrder {
		_, err = tx.Exec(r.Context(), `
			UPDATE sections
			SET "order" = "order" + 1
			WHERE challenge_id = $1 AND "order" >= $2 AND "order" < $3
		`, challengeID, req.Order, currentOrder)
	} else if req.Order > currentOrder {
		// If moving down (larger order number)
		_, err = tx.Exec(r.Context(), `
			UPDATE sections
			SET "order" = "order" - 1
			WHERE challenge_id = $1 AND "order" > $2 AND "order" <= $3
		`, challengeID, currentOrder, req.Order)
	} else {
		// No change needed
		utils.Success(w, http.StatusOK, map[string]string{"message": "Order unchanged"})
		return
	}

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update order of other sections")
		return
	}

	// Update the section's order
	_, err = tx.Exec(r.Context(), `UPDATE sections SET "order" = $1 WHERE id = $2`, req.Order, sectionID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update section order")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]string{"message": "Section reordered successfully"})
}

// Helper function to validate challenge ownership
func validateChallengeOwnership(challengeID string, userID string) error {
	// Parse UUID
	_, err := uuid.Parse(challengeID)
	if err != nil {
		return err
	}

	// Check if challenge exists and belongs to user
	var count int
	err = db.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM challenges WHERE id = $1 AND user_id = (SELECT id FROM users WHERE clerk_id = $2)
	`, challengeID, userID).Scan(&count)

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("challenge not found or doesn't belong to user")
	}

	return nil
}

// Helper function to validate section ownership
func validateSectionOwnership(sectionID string, userID string) error {
	// Parse UUID
	_, err := uuid.Parse(sectionID)
	if err != nil {
		return err
	}

	// Check if section exists and belongs to user's challenge
	var count int
	err = db.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM sections s
		JOIN challenges c ON s.challenge_id = c.id
		JOIN users u ON c.user_id = u.id
		WHERE s.id = $1 AND u.clerk_id = $2
	`, sectionID, userID).Scan(&count)

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("section not found or doesn't belong to user's challenge")
	}

	return nil
}
