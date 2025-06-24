package models

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
    Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ApiResponseTwo struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
    Error   string `json:"error,omitempty"`
}
