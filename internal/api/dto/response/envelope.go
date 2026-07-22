package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope adalah standar response body untuk seluruh endpoint.
type Envelope struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK mengembalikan envelope sukses dengan data dan pesan tertentu.
func OK(message string, data any) Envelope {
	return Envelope{Status: "ok", Message: message, Data: data}
}

// ErrorEnvelope mengembalikan envelope gagal dengan pesan tertentu.
func ErrorEnvelope(message string) Envelope {
	return Envelope{Status: "error", Message: message, Data: nil}
}

// WriteOK menulis response JSON sukses (200) dengan envelope standar.
func WriteOK(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusOK, OK(message, data))
}

// WriteCreated menulis response JSON sukses (201) dengan envelope standar.
func WriteCreated(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusCreated, OK(message, data))
}

// WriteError menulis response JSON gagal dengan status code tertentu.
func WriteError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, ErrorEnvelope(message))
}
