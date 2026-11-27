package inquiry

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

func validRequest() domain.InquiryRequest {
	return domain.InquiryRequest{
		MosqueName: "Masjid Al-Ikhlas",
		Plan:       domain.InquiryPlanGratis,
		Subdomain:  "alikhlas",
		FullName:   "Ahmad Fauzi",
		Phone:      "081234567890",
		Domicile:   "Banjarbaru",
		Source:     "Media sosial",
		Notes:      "Ada 2 TV, satu di serambi.",
		Consent:    true,
	}
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		return appErr.Code
	}
	return ""
}

// fakeNotifier records what it was asked to send
type fakeNotifier struct {
	sent []string
	err  error
}

func (f *fakeNotifier) Send(_ context.Context, text string) error {
	f.sent = append(f.sent, text)
	return f.err
}

func newUsecase(envTag string) (*UseCase, *fakeNotifier) {
	notifier := &fakeNotifier{}
	s := New(notifier, envTag)
	s.now = func() time.Time { return time.Date(2026, 11, 19, 7, 3, 0, 0, time.UTC) }
	return s, notifier
}

func TestSubmit(t *testing.T) {
	body := "📩 <b>Pendaftaran baru</b>\n\n" +
		"🕌 <b>Masjid Al-Ikhlas</b>\n" +
		"Paket: <b>Gratis</b>\n" +
		"Alamat: <code>alikhlas.masjidyuk.com</code>\n\n" +
		"👤 <b>Ahmad Fauzi</b>\n" +
		"WhatsApp: <a href=\"https://wa.me/6281234567890\">+6281234567890</a>\n" +
		"Kota: Banjarbaru\n" +
		"Tahu dari: Media sosial\n\n" +
		"📝 <b>Catatan</b>\n<blockquote>Ada 2 TV, satu di serambi.</blockquote>\n\n" +
		"<i>Dikirim 19 Nov 2026 14:03 WIB</i>"
	tests := []struct{ name, envTag, want string }{
		{"production has no tag", "", body},
		{"staging is tagged", " staging ", "<code>STAGING</code> " + body},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, notifier := newUsecase(tt.envTag)
			if err := s.Submit(context.Background(), validRequest()); err != nil {
				t.Fatalf("Submit: %v", err)
			}
			if len(notifier.sent) != 1 || notifier.sent[0] != tt.want {
				t.Errorf("sent = %q\nwant %q", notifier.sent, tt.want)
			}
		})
	}
}

func TestSubmitDonatur(t *testing.T) {
	s, notifier := newUsecase("")
	req := validRequest()
	req.Plan = domain.InquiryPlanDonatur
	req.Donation = 12_500_000
	req.Source, req.Notes = "", ""

	if err := s.Submit(context.Background(), req); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	want := "📩 <b>Pendaftaran baru</b>\n\n" +
		"🕌 <b>Masjid Al-Ikhlas</b>\n" +
		"Paket: <b>Donatur</b>\n" +
		"Domain: <i>belum ditentukan</i>\n" +
		"Donasi: <b>Rp12.500.000</b> / tahun\n\n" +
		"👤 <b>Ahmad Fauzi</b>\n" +
		"WhatsApp: <a href=\"https://wa.me/6281234567890\">+6281234567890</a>\n" +
		"Kota: Banjarbaru\n\n" +
		"<i>Dikirim 19 Nov 2026 14:03 WIB</i>"
	if len(notifier.sent) != 1 || notifier.sent[0] != want {
		t.Errorf("sent = %q\nwant %q", notifier.sent, want)
	}
}

func TestSubmitEscapesWhatVisitorsType(t *testing.T) {
	s, notifier := newUsecase("")
	req := validRequest()
	req.MosqueName = "Masjid <b>Al</b> & Co"
	req.Notes = "</blockquote><a href=\"x\">klik</a>"

	if err := s.Submit(context.Background(), req); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	sent := notifier.sent[0]
	for _, want := range []string{
		"🕌 <b>Masjid &lt;b&gt;Al&lt;/b&gt; &amp; Co</b>",
		"<blockquote>&lt;/blockquote&gt;&lt;a href=&#34;x&#34;&gt;klik&lt;/a&gt;</blockquote>",
	} {
		if !strings.Contains(sent, want) {
			t.Errorf("sent = %q\nwant it to contain %q", sent, want)
		}
	}
}

func TestRupiah(t *testing.T) {
	for in, want := range map[int64]string{0: "Rp0", 999: "Rp999", 1_000_000: "Rp1.000.000", 12_500_000: "Rp12.500.000"} {
		if got := rupiah(in); got != want {
			t.Errorf("rupiah(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestSubmitWithoutSource(t *testing.T) {
	s, notifier := newUsecase("")
	req := validRequest()
	req.Source = ""
	if err := s.Submit(context.Background(), req); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if len(notifier.sent) != 1 || strings.Contains(notifier.sent[0], "Tahu dari") {
		t.Errorf("sent = %q, want no Tahu dari line", notifier.sent)
	}
}

func TestSubmitErrors(t *testing.T) {
	t.Run("invalid input is not sent", func(t *testing.T) {
		s, notifier := newUsecase("")
		req := validRequest()
		req.Consent = false
		if err := s.Submit(context.Background(), req); codeOf(err) != apperror.CodeInvalidInput || len(notifier.sent) != 0 {
			t.Errorf("Submit = %v, sent %d; want INVALID_INPUT and nothing sent", err, len(notifier.sent))
		}
	})

	t.Run("bots get no error and nothing is sent", func(t *testing.T) {
		s, notifier := newUsecase("")
		req := validRequest()
		req.Website = "http://spam.example"
		if err := s.Submit(context.Background(), req); err != nil || len(notifier.sent) != 0 {
			t.Errorf("Submit = %v, sent %d; want nil and nothing", err, len(notifier.sent))
		}
	})

	t.Run("send fails", func(t *testing.T) {
		s, notifier := newUsecase("")
		notifier.err = errors.New("telegram: status 502")
		if err := s.Submit(context.Background(), validRequest()); codeOf(err) != apperror.CodeUnavailable {
			t.Errorf("Submit error = %v, want UNAVAILABLE", err)
		}
	})

	t.Run("not configured", func(t *testing.T) {
		if err := New(nil, "").Submit(context.Background(), validRequest()); codeOf(err) != apperror.CodeUnavailable {
			t.Errorf("Submit error = %v, want UNAVAILABLE", err)
		}
	})
}
