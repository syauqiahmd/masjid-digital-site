package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http/controller"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

type fakeInquiries struct {
	got domain.InquiryRequest
	err error
}

func (f *fakeInquiries) Submit(_ context.Context, req domain.InquiryRequest) error {
	f.got = req
	return f.err
}

type fakeReports struct {
	report domain.Transparency
	err    error
}

func (f *fakeReports) Get(context.Context) (domain.Transparency, error) { return f.report, f.err }

func do(h http.Handler, method, path, body, remote string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if remote != "" {
		req.RemoteAddr = remote
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSubmitInquiry(t *testing.T) {
	inquiries := &fakeInquiries{}
	h := newRouter(inquiries, &fakeReports{})

	body := `{"mosque_name":"Masjid Al-Ikhlas","plan":"donatur","domain":"alikhlas.id","donation":1500000,
		"full_name":"Ahmad Fauzi","phone":"081234567890","domicile":"Banjarbaru","source":"Media sosial","notes":"Dua TV","consent":true}`
	rec := do(h, http.MethodPost, "/api/public/v1/inquiries", body, "")
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201: %s", rec.Code, rec.Body)
	}
	want := domain.InquiryRequest{
		MosqueName: "Masjid Al-Ikhlas", Plan: domain.InquiryPlanDonatur, Domain: "alikhlas.id", Donation: 1_500_000,
		FullName: "Ahmad Fauzi", Phone: "081234567890", Domicile: "Banjarbaru", Source: "Media sosial", Notes: "Dua TV", Consent: true,
	}
	if inquiries.got != want {
		t.Errorf("submitted %+v, want %+v", inquiries.got, want)
	}
}

func TestSubmitInquiryErrors(t *testing.T) {
	tests := []struct {
		name, body string
		err        error
		status     int
		want       string
	}{
		{"malformed JSON", `{"full_name":`, nil, 400, `{"code":"INVALID_INPUT","message":"format request tidak valid"}`},
		{"invalid input", `{}`, apperror.New(apperror.CodeInvalidInput, errors.New("missing"), "nama lengkap wajib diisi"), 400, `{"code":"INVALID_INPUT","message":"nama lengkap wajib diisi"}`},
		{"chat down", `{}`, apperror.New(apperror.CodeUnavailable, errors.New("telegram: 502"), "pesan belum terkirim, silakan coba lagi beberapa saat lagi"), 503, `{"code":"UNAVAILABLE","message":"pesan belum terkirim, silakan coba lagi beberapa saat lagi"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newRouter(&fakeInquiries{err: tt.err}, &fakeReports{})
			rec := do(h, http.MethodPost, "/api/public/v1/inquiries", tt.body, "")
			if rec.Code != tt.status || strings.TrimSpace(rec.Body.String()) != tt.want {
				t.Errorf("got %d %s, want %d %s", rec.Code, rec.Body, tt.status, tt.want)
			}
		})
	}
}

func TestSubmitInquiryRateLimit(t *testing.T) {
	h := newRouter(&fakeInquiries{}, &fakeReports{})

	for i := range inquiryBurstPerClient {
		if rec := do(h, http.MethodPost, "/api/public/v1/inquiries", `{}`, "203.0.113.7:5000"); rec.Code != http.StatusCreated {
			t.Fatalf("request %d: status %d", i+1, rec.Code)
		}
	}
	if rec := do(h, http.MethodPost, "/api/public/v1/inquiries", `{}`, "203.0.113.7:5000"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("past the burst: status %d, want 429", rec.Code)
	}
	if rec := do(h, http.MethodPost, "/api/public/v1/inquiries", `{}`, "203.0.113.8:5000"); rec.Code != http.StatusCreated {
		t.Errorf("another client: status %d, want 201", rec.Code)
	}
}

func TestGetTransparency(t *testing.T) {
	safeUntil := "2028-03-14"
	h := newRouter(&fakeInquiries{}, &fakeReports{report: domain.Transparency{
		Deposit: 18_750_000, MonthlyCost: 1_200_000, SafeUntil: &safeUntil,
		UpdatedAt: time.Date(2026, 11, 27, 7, 0, 0, 0, time.UTC),
	}})

	rec := do(h, http.MethodGet, "/api/public/v1/transparency", "", "")
	want := `{"data":{"deposit":18750000,"monthly_cost":1200000,"safe_until":"2028-03-14","updated_at":"2026-11-27T07:00:00Z"}}`
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != want {
		t.Errorf("got %d %s\nwant 200 %s", rec.Code, rec.Body, want)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=300" {
		t.Errorf("Cache-Control = %q", got)
	}
}

func TestGetTransparencyUnavailable(t *testing.T) {
	h := newRouter(&fakeInquiries{}, &fakeReports{err: apperror.New(apperror.CodeUnavailable, errors.New("down"), "laporan server belum tersedia")})
	rec := do(h, http.MethodGet, "/api/public/v1/transparency", "", "")
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"UNAVAILABLE"`) {
		t.Errorf("got %d %s, want 503 UNAVAILABLE", rec.Code, rec.Body)
	}
}

func TestUnknownRoute(t *testing.T) {
	h := newRouter(&fakeInquiries{}, &fakeReports{})
	if rec := do(h, http.MethodGet, "/api/public/v1/nope", "", ""); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestUnexpectedErrorIs500(t *testing.T) {
	h := newRouter(&fakeInquiries{err: errors.New("boom")}, &fakeReports{})
	rec := do(h, http.MethodPost, "/api/public/v1/inquiries", `{}`, "")
	want := `{"code":"INTERNAL","message":"terjadi kesalahan, silakan coba lagi"}`
	if rec.Code != http.StatusInternalServerError || strings.TrimSpace(rec.Body.String()) != want {
		t.Errorf("got %d %s, want 500 %s", rec.Code, rec.Body, want)
	}
}

func newRouter(inquiries controller.InquiryUsecaseItf, reports controller.TransparencyUsecaseItf) http.Handler {
	return NewRouter(controller.NewController(inquiries, reports))
}
