package httputil

import "github.com/labstack/echo/v5"

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
}

type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

func Success(c *echo.Context, code int, data interface{}) error {
	return (*c).JSON(code, Response{Success: true, Data: data})
}

func SuccessMsg(c *echo.Context, code int, msg string) error {
	return (*c).JSON(code, Response{Success: true, Message: msg})
}

func SuccessWithPagination(c *echo.Context, data interface{}, total, limit, offset int) error {
	return (*c).JSON(200, PaginatedResponse{
		Success: true,
		Data:    data,
		Meta:    PaginationMeta{Total: total, Limit: limit, Offset: offset},
	})
}

func Error(c *echo.Context, code int, msg, errCode string) error {
	return (*c).JSON(code, Response{Success: false, Message: msg, ErrorCode: errCode})
}
