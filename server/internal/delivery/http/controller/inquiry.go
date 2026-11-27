package controller

import (
	"encoding/json"
	"net/http"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

const maxInquiryBodySize = 16 << 10

// SubmitInquiry handles POST /inquiries, the Daftar form. The router rate limits it, since anyone
// can call it.
func (ctrl *Controller) SubmitInquiry(w http.ResponseWriter, r *http.Request) error {
	var req domain.InquiryRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxInquiryBodySize)).Decode(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	if err := ctrl.inquiry.Submit(r.Context(), req); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, messageResponse{Message: "terima kasih, pesan Anda sudah kami terima"})
}
