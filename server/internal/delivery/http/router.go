// Package http serves the public site's API: the Daftar form and the "server bersama" numbers.
// The site's Caddy forwards /api/* here.
package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http/controller"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http/middleware"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

// The inquiry form is public and each submission pings the team on Telegram, so it is rate
// limited per client and in total
const (
	inquiryBurstPerClient    = 5
	inquiryIntervalPerClient = 10 * time.Minute
	inquiryBurstTotal        = 30
	inquiryIntervalTotal     = 2 * time.Minute
)

// NewRouter routes /api/public/v1/* to ctrl, plus GET /healthz for the container's health check
func NewRouter(ctrl *controller.Controller) http.Handler {
	perClient := middleware.RateLimit(inquiryIntervalPerClient, inquiryBurstPerClient, middleware.ByClientIP)
	total := middleware.RateLimit(inquiryIntervalTotal, inquiryBurstTotal, middleware.Everyone)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// public, no login
	mux.HandleFunc("POST /api/public/v1/inquiries", handle(perClient(total(ctrl.SubmitInquiry))))
	mux.HandleFunc("GET /api/public/v1/transparency", handle(ctrl.GetTransparency))

	mux.HandleFunc("/api/", handle(func(http.ResponseWriter, *http.Request) error {
		return apperror.New(apperror.CodeNotFound, errors.New("no such route"), "tidak ditemukan")
	}))

	return mux
}
