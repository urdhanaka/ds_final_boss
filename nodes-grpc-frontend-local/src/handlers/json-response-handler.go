package handlers

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewSuccessResponse() SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: "success",
		Data:    nil,
	}
}

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "request error",
		Error:   err.Error(),
	}
}

func NewSuccessResponseWithData(data any) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: "success",
		Data:    data,
	}
}
