package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JSON 返回方式
type JSONBuilder struct {
	c      *gin.Context
	status int
	code   int
	msg    string
	data   any
	page   *Pagination
}

// JSON creates a new JSONBuilder bound to the given Gin context.
func JSON(c *gin.Context) *JSONBuilder {
	return &JSONBuilder{c: c}
}

// Status sets the HTTP status code independently from the business code.
func (b *JSONBuilder) Status(status int) *JSONBuilder {
	b.status = status
	return b
}

// Code sets the business / HTTP code.
// If no explicit HTTP status is set, it will also be used as status.
func (b *JSONBuilder) Code(code int) *JSONBuilder {
	b.code = code
	if b.status == 0 {
		b.status = code
	}
	return b
}

// Message sets the response message.
func (b *JSONBuilder) Message(msg string) *JSONBuilder {
	b.msg = msg
	return b
}

// Data sets the response data payload.
func (b *JSONBuilder) Data(data any) *JSONBuilder {
	b.data = data
	return b
}

// Pagination sets pagination info.
func (b *JSONBuilder) Pagination(page, pageSize, total int) *JSONBuilder {
	b.page = &Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	return b
}

// Send finalizes and writes the response.
func (b *JSONBuilder) Send() {
	// Defaults
	if b.status == 0 {
		if b.code != 0 {
			b.status = b.code
		} else {
			b.status = http.StatusOK
		}
	}
	if b.code == 0 {
		b.code = b.status
	}

	// traceID := utils.TraceIDFromContext(b.c.Request.Context())
	resp := Response{
		Code:       b.code,
		Message:    b.msg,
		Data:       b.data,
		Pagination: b.page,
		// TraceID:    traceID,
	}
	b.c.JSON(b.status, resp)
}
