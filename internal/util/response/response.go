package response

import "github.com/gin-gonic/gin"

type SuccessResponse struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func SuccessResponseUtil(code int, message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
		Code:    code,
		Data:    data,
	}
}

func ErrorResponseUtil(code int, message string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: message,
		Code:    code,
	}
}

func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, SuccessResponseUtil(code, message, data))
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponseUtil(code, message))
}
