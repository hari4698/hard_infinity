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

type CreateTaskRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	TaskType      string `json:"task_type"`
	Required      bool   `json:"required"`
	RestartOnFail bool   `json:"restart_on_fail"`
	StrikesEnabled bool   `json:"strikes_enabled"`
	StrikesLimit  int    `json:"strikes_limit"`
	Order         int    `json:"order"`
}

type UpdateTaskRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	TaskType      string `json:"task_type"`
	Required      bool   `json:"required"`
	RestartOnFail bool   `json:"restart_on_fail"`
	StrikesEnabled bool   `json:"strikes_enabled"`
	StrikesLimit  int    `json:"strikes_limit"`
}

type ReorderTaskRequest struct {
	Order int `json:"order"`
}

// GetTasks retrieves all tasks for a section
func GetTasks(w http.ResponseWriter, r *http.Request) {
	sectionID := chi.URLParam(r, "id")

	// Verify that the section exists and belongs to the authenticated user
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check section ownership
	if err := validateSectionOwnership(sectionID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Section not found")
		return
	}

	// Query the database for tasks
	rows, err := db.DB.Query(r.Context(), `
		SELECT id, section_id, name, description, task_type, required, restart_on_fail, 
		       strikes_enabled, strikes_limit, "order", created_at, updated_at
		FROM tasks
		WHERE section_id = $1
		ORDER BY "order" ASC
	`, sectionID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve tasks")
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(
			&task.ID,
			&task.SectionID,
			&task.Name,
			&task.Description,
			&task.TaskType,
			&task.Required,
			&task.RestartOnFail,
			&task.StrikesEnabled,
			&task.StrikesLimit,
			&task.Order,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Error scanning task data")
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Error iterating through tasks")
		return
	}

	utils.Success(w, http.StatusOK, tasks)
}

// CreateTask creates a new task for a section
func CreateTask(w http.ResponseWriter, r *http.Request) {
	sectionID := chi.URLParam(r, "id")

	// Verify that the section exists and belongs to the authenticated user
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check section ownership
	if err := validateSectionOwnership(sectionID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Section not found")
		return
	}

	// Parse request body
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		utils.Error(w, http.StatusBadRequest, "Task name is required")
		return
	}

	if req.TaskType == "" {
		utils.Error(w, http.StatusBadRequest, "Task type is required")
		return
	}

	// Set default values
	if req.StrikesEnabled && req.StrikesLimit <= 0 {
		req.StrikesLimit = 3 // Default to 3 strikes if enabled but no limit specified
	}

	// Generate a new UUID for the task
	taskID := uuid.New().String()

	// If order is not specified, determine the next available order
	if req.Order <= 0 {
		var maxOrder int
		err := db.DB.QueryRow(r.Context(), `
			SELECT COALESCE(MAX("order"), 0) FROM tasks WHERE section_id = $1
		`, sectionID).Scan(&maxOrder)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to determine task order")
			return
		}
		req.Order = maxOrder + 1
	}

	// Insert the new task
	var task models.Task
	err := db.DB.QueryRow(r.Context(), `
		INSERT INTO tasks (id, section_id, name, description, task_type, required, restart_on_fail, 
		                  strikes_enabled, strikes_limit, "order", created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, section_id, name, description, task_type, required, restart_on_fail, 
		          strikes_enabled, strikes_limit, "order", created_at, updated_at
	`, taskID, sectionID, req.Name, req.Description, req.TaskType, req.Required, req.RestartOnFail,
		req.StrikesEnabled, req.StrikesLimit, req.Order).Scan(
		&task.ID,
		&task.SectionID,
		&task.Name,
		&task.Description,
		&task.TaskType,
		&task.Required,
		&task.RestartOnFail,
		&task.StrikesEnabled,
		&task.StrikesLimit,
		&task.Order,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	utils.Success(w, http.StatusCreated, task)
}

// UpdateTask updates an existing task
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	// Verify user is authenticated
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check task ownership through section and challenge
	if err := validateTaskOwnership(taskID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Task not found")
		return
	}

	// Parse request body
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		utils.Error(w, http.StatusBadRequest, "Task name is required")
		return
	}

	if req.TaskType == "" {
		utils.Error(w, http.StatusBadRequest, "Task type is required")
		return
	}

	// Set default values
	if req.StrikesEnabled && req.StrikesLimit <= 0 {
		req.StrikesLimit = 3 // Default to 3 strikes if enabled but no limit specified
	}

	// Update the task
	var task models.Task
	err := db.DB.QueryRow(r.Context(), `
		UPDATE tasks
		SET name = $1, description = $2, task_type = $3, required = $4, restart_on_fail = $5,
		    strikes_enabled = $6, strikes_limit = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING id, section_id, name, description, task_type, required, restart_on_fail, 
		          strikes_enabled, strikes_limit, "order", created_at, updated_at
	`, req.Name, req.Description, req.TaskType, req.Required, req.RestartOnFail,
		req.StrikesEnabled, req.StrikesLimit, taskID).Scan(
		&task.ID,
		&task.SectionID,
		&task.Name,
		&task.Description,
		&task.TaskType,
		&task.Required,
		&task.RestartOnFail,
		&task.StrikesEnabled,
		&task.StrikesLimit,
		&task.Order,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	utils.Success(w, http.StatusOK, task)
}

// DeleteTask deletes a task
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	// Verify user is authenticated
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check task ownership through section and challenge
	if err := validateTaskOwnership(taskID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Task not found")
		return
	}

	// Get the section ID and current order for reordering remaining tasks
	var sectionID string
	var currentOrder int
	err := db.DB.QueryRow(r.Context(), `
		SELECT section_id, "order" FROM tasks WHERE id = $1
	`, taskID).Scan(&sectionID, &currentOrder)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get task info")
		return
	}

	// Start a transaction to delete the task and update orders
	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	// Delete the task
	_, err = tx.Exec(r.Context(), "DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	// Update order of remaining tasks
	_, err = tx.Exec(r.Context(), `
		UPDATE tasks
		SET "order" = "order" - 1
		WHERE section_id = $1 AND "order" > $2
	`, sectionID, currentOrder)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update task orders")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}

func ReorderTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	// Verify user is authenticated
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check task ownership through section and challenge
	if err := validateTaskOwnership(taskID, userID); err != nil {
		utils.Error(w, http.StatusNotFound, "Task not found")
		return
	}

	// Parse request body
	var req ReorderTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Order <= 0 {
		utils.Error(w, http.StatusBadRequest, "Order must be a positive integer")
		return
	}

	// Get the task's section ID and current order
	var sectionID string
	var currentOrder int
	err := db.DB.QueryRow(r.Context(), `
		SELECT section_id, "order" FROM tasks WHERE id = $1
	`, taskID).Scan(&sectionID, &currentOrder)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get task info")
		return
	}

	// Update the task's order
	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	// If moving up (smaller order number)
	if req.Order < currentOrder {
		_, err = tx.Exec(r.Context(), `
			UPDATE tasks
			SET "order" = "order" + 1
			WHERE section_id = $1 AND "order" >= $2 AND "order" < $3
		`, sectionID, req.Order, currentOrder)
	} else if req.Order > currentOrder {
		// If moving down (larger order number)
		_, err = tx.Exec(r.Context(), `
			UPDATE tasks
			SET "order" = "order" - 1
			WHERE section_id = $1 AND "order" > $2 AND "order" <= $3
		`, sectionID, currentOrder, req.Order)
	} else {
		// No change needed
		utils.Success(w, http.StatusOK, map[string]string{"message": "Order unchanged"})
		return
	}

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update order of other tasks")
		return
	}

	// Update the task's order
	_, err = tx.Exec(r.Context(), `
		UPDATE tasks SET "order" = $1 WHERE id = $2
	`, req.Order, taskID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update task order")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.Success(w, http.StatusOK, map[string]string{"message": "Task reordered successfully"})
}

func validateTaskOwnership(taskID string, userID string) error {
	// Parse UUID
	_, err := uuid.Parse(taskID)
	if err != nil {
		return err
	}

	// Check if task exists and belongs to user's challenge
	var count int
	err = db.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM tasks t
		JOIN sections s ON t.section_id = s.id
		JOIN challenges c ON s.challenge_id = c.id
		JOIN users u ON c.user_id = u.id
		WHERE t.id = $1 AND u.clerk_id = $2
	`, taskID, userID).Scan(&count)

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("task not found or doesn't belong to user's challenge")
	}

	return nil
}