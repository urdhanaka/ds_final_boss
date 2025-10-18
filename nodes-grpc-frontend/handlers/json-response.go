package handlers

type SuccessResponse struct {
	Success bool
	Message string
	Data    any
}

type ErrorResponse struct {
	Success bool
	Message string
	Error   string
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
