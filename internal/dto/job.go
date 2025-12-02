package dto

type SubmitCodeBody struct {
	Code     string  `json:"code" binding:"required"`
	Language string  `json:"language" binding:"required"`
	Input    *string `json:"input"`
}
