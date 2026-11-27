package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http/controller"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http/middleware"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

var statusOf = map[apperror.Code]int{
	apperror.CodeInvalidInput:    http.StatusBadRequest,
	apperror.CodeNotFound:        http.StatusNotFound,
	apperror.CodeTooManyRequests: http.StatusTooManyRequests,
	apperror.CodeInternal:        http.StatusInternalServerError,
	apperror.CodeUnavailable:     http.StatusServiceUnavailable,
}

type errorResponse struct {
	Code    apperror.Code `json:"code"`
	Message string        `json:"message"`
}

// handle adapts a handler that returns its error: an *apperror.AppError becomes its status and
// {"code", "message"}, anything else a 500 whose cause is only logged
func handle(h middleware.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		appErr, ok := errors.AsType[*apperror.AppError](err)
		if !ok {
			appErr = apperror.New(apperror.CodeInternal, err, "terjadi kesalahan, silakan coba lagi")
		}
		status, ok := statusOf[appErr.Code]
		if !ok {
			status = http.StatusInternalServerError
		}
		if status >= http.StatusInternalServerError {
			log.Printf("%s %s: %v", r.Method, r.URL.Path, appErr)
		}

		if err := controller.WriteJSON(w, status, errorResponse{Code: appErr.Code, Message: appErr.MessageError}); err != nil {
			log.Printf("write error response: %v", err)
		}
	}
}
