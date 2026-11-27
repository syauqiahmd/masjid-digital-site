// Package controller turns HTTP requests into usecase calls. Errors are returned to the server's
// error handler, which writes {"code", "message"}.
package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
)

type InquiryUsecaseItf interface {
	Submit(ctx context.Context, req domain.InquiryRequest) error
}

type TransparencyUsecaseItf interface {
	Get(ctx context.Context) (domain.Transparency, error)
}

type Controller struct {
	inquiry      InquiryUsecaseItf
	transparency TransparencyUsecaseItf
}

func NewController(inquiry InquiryUsecaseItf, transparency TransparencyUsecaseItf) *Controller {
	return &Controller{inquiry: inquiry, transparency: transparency}
}

// dataResponse wraps a successful payload as {"data": ...}
type dataResponse struct {
	Data any `json:"data"`
}

type messageResponse struct {
	Message string `json:"message"`
}

// WriteJSON writes body as JSON with the status
func WriteJSON(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(body)
}
