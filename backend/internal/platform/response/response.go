package response

import (
	"encoding/json"
	"net/http"
)

type envelope struct {
	Data interface{} `json:"data,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type paginatedEnvelope struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data})
}

func Paginated(w http.ResponseWriter, data interface{}, p Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paginatedEnvelope{Data: data, Pagination: p})
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorBody{Code: code, Message: message},
	})
}

func ValidationError(w http.ResponseWriter, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: details,
		},
	})
}
