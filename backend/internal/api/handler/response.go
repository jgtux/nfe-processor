package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API envelope for all endpoints.
type Response[T any] struct {
	Data  T      `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Total int    `json:"total,omitempty"`
}

// OK responds with a single object wrapped in the envelope.
func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Response[T]{Data: data})
}

// OKList responds with a slice and its count.
func OKList[T any](c *gin.Context, data []T) {
	c.JSON(http.StatusOK, Response[[]T]{Data: data, Total: len(data)})
}

// Created responds with 202 Accepted and a payload.
func Created[T any](c *gin.Context, data T) {
	c.JSON(http.StatusAccepted, Response[T]{Data: data})
}

// Fail responds with an error envelope.
func Fail(c *gin.Context, status int, err error) {
	c.JSON(status, Response[any]{Error: err.Error()})
}
