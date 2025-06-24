package handlers

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func NewSuccessResponse(message string) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
		Data:    nil,
	}
}

func NewErrorResponse(err error, message string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: message,
		Error:   err.Error(),
	}
}

func NewSuccessResponseWithData(data any, message string) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: "success",
		Data:    data,
	}
}
