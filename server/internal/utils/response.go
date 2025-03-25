package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`	
}

// JSON sends a JSON response with appropriate headers
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Success sends a success response
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	response := Response{
		Success: true,
		Data:    data,
	}
	JSON(w, statusCode, response)
}

// Error sends an error response
func Error(w http.ResponseWriter, statusCode int, err string) {
	response := Response{
		Success: false,
		Error:   err,
	}
	JSON(w, statusCode, response)
}