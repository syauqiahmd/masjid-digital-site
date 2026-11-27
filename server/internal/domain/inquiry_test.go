package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

func validInquiryRequest() InquiryRequest {
	return InquiryRequest{
		MosqueName: "Masjid Al-Ikhlas",
		Plan:       InquiryPlanGratis,
		Subdomain:  "alikhlas",
		FullName:   "Ahmad Fauzi",
		Phone:      "081234567890",
		Domicile:   "Banjarbaru",
		Source:     "Media sosial",
		Consent:    true,
	}
}

func TestInquiryRequestValidate(t *testing.T) {
	r := InquiryRequest{
		MosqueName: "  Masjid   Al-Ikhlas ",
		Plan:       " Gratis ",
		Subdomain:  " AlIkhlas ",
		Domain:     "ignored.id", // the other plan's field
		Donation:   5_000_000,
		FullName:   "  Ahmad   Fauzi ",
		Phone:      "0812-3456-7890",
		Domicile:   "Banjarbaru\nKalimantan Selatan",
		Source:     "Media sosial",
		Notes:      "  Kami punya 2 TV.\nSatu di serambi.  ",
		Consent:    true,
	}
	got, err := r.Validate()
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}

	want := Inquiry{
		MosqueName: "Masjid Al-Ikhlas",
		Plan:       InquiryPlanGratis,
		Subdomain:  "alikhlas",
		FullName:   "Ahmad Fauzi",
		Phone:      "+6281234567890",
		Domicile:   "Banjarbaru Kalimantan Selatan",
		Source:     "Media sosial",
		Notes:      "Kami punya 2 TV.\nSatu di serambi.",
	}
	if got != want {
		t.Errorf("Validate = %+v\nwant %+v", got, want)
	}
	if url := got.WhatsAppURL(); url != "https://wa.me/6281234567890" {
		t.Errorf("WhatsAppURL = %q", url)
	}
}

func TestInquiryRequestValidateDonatur(t *testing.T) {
	r := validInquiryRequest()
	r.Plan = InquiryPlanDonatur
	r.Domain = " AlIkhlas.OR.id. "
	r.Donation = 1_500_000

	got, err := r.Validate()
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if got.Domain != "alikhlas.or.id" || got.Donation != 1_500_000 || got.Subdomain != "" {
		t.Errorf("Validate = %+v", got)
	}
}

func TestInquiryRequestValidateRejects(t *testing.T) {
	donatur := func(r *InquiryRequest) { r.Plan = InquiryPlanDonatur; r.Donation = MinDonation }
	tests := []struct {
		name    string
		change  func(*InquiryRequest)
		message string
	}{
		{"no mosque", func(r *InquiryRequest) { r.MosqueName = "  " }, "nama masjid wajib diisi"},
		{"no plan", func(r *InquiryRequest) { r.Plan = "" }, "paket wajib diisi"},
		{"unknown plan", func(r *InquiryRequest) { r.Plan = "premium" }, "paket tidak dikenal"},
		{"no name", func(r *InquiryRequest) { r.FullName = "  " }, "nama pengurus wajib diisi"},
		{"no phone", func(r *InquiryRequest) { r.Phone = "" }, "nomor WhatsApp wajib diisi"},
		{"bad phone", func(r *InquiryRequest) { r.Phone = "12345" }, "nomor WhatsApp tidak valid"},
		{"landline", func(r *InquiryRequest) { r.Phone = "0215551234" }, "nomor WhatsApp tidak valid"},
		{"no city", func(r *InquiryRequest) { r.Domicile = "\n" }, "kota wajib diisi"},
		{"long mosque", func(r *InquiryRequest) { r.MosqueName = strings.Repeat("a", 101) }, "nama masjid maksimal 100 karakter"},
		{"long source", func(r *InquiryRequest) { r.Source = strings.Repeat("a", 101) }, "informasi lainnya maksimal 100 karakter"},
		{"long notes", func(r *InquiryRequest) { r.Notes = strings.Repeat("a", 1001) }, "catatan maksimal 1000 karakter"},
		{"bad subdomain", func(r *InquiryRequest) { r.Subdomain = "al_ikhlas" }, "alamat hanya boleh huruf kecil, angka, dan tanda hubung, 3 sampai 63 karakter"},
		{"short subdomain", func(r *InquiryRequest) { r.Subdomain = "ab" }, "alamat hanya boleh huruf kecil, angka, dan tanda hubung, 3 sampai 63 karakter"},
		{"bad domain", func(r *InquiryRequest) { donatur(r); r.Domain = "masjid kamu" }, "domain tidak valid, contoh: masjidkamu.id"},
		{"domain without TLD", func(r *InquiryRequest) { donatur(r); r.Domain = "alikhlas" }, "domain tidak valid, contoh: masjidkamu.id"},
		{"donation too small", func(r *InquiryRequest) { donatur(r); r.Donation = 999_999 }, "donasi minimal Rp1.000.000 per tahun untuk paket Donatur"},
		{"no consent", func(r *InquiryRequest) { r.Consent = false }, "centang persetujuan untuk dihubungi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := validInquiryRequest()
			tt.change(&r)
			_, err := r.Validate()
			appErr, ok := errors.AsType[*apperror.AppError](err)
			if !ok || appErr.Code != apperror.CodeInvalidInput || appErr.MessageError != tt.message {
				t.Errorf("Validate error = %v, want %q", err, tt.message)
			}
		})
	}
}

func TestNormalizePhone(t *testing.T) {
	for in, want := range map[string]string{
		"0812 3456 7890":   "+6281234567890",
		"6281234567890":    "+6281234567890",
		"+62 812-3456-789": "+628123456789",
		"+60123456789":     "+60123456789",
	} {
		if got, ok := normalizePhone(in); !ok || got != want {
			t.Errorf("normalizePhone(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
}

func TestInquiryRequestIsBot(t *testing.T) {
	r := validInquiryRequest()
	if r.IsBot() {
		t.Error("IsBot = true without the honeypot")
	}
	r.Website = "http://spam.example"
	if !r.IsBot() {
		t.Error("IsBot = false with the honeypot filled in")
	}
}
