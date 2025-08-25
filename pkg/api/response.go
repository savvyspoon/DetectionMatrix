package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse defines the standard error payload for API responses.
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      int    `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

// ListResponse is the standard envelope for list endpoints.
type ListResponse struct {
	Items    interface{} `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
}

// List writes a standardized list envelope with OK status.
func List(w http.ResponseWriter, items interface{}, page, pageSize, total int) {
	JSON(w, http.StatusOK, ListResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

// JSON writes a JSON response with the given status code and payload.
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Error writes a standardized JSON error response.
func Error(w http.ResponseWriter, r *http.Request, status int, message string) {
	reqID := ""
	if v := r.Context().Value(RequestIDKey); v != nil {
		if id, ok := v.(string); ok {
			reqID = id
		}
	}
	JSON(w, status, ErrorResponse{Error: message, Code: status, RequestID: reqID})
}
