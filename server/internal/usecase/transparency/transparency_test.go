package transparency

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

// fakeBilling answers each call with the next of its answers
type fakeBilling struct {
	answers []fakeAnswer
	calls   int
}

type fakeAnswer struct {
	billing domain.HostingBilling
	err     error
}

func (f *fakeBilling) Billing(context.Context) (domain.HostingBilling, error) {
	if f.calls >= len(f.answers) {
		return domain.HostingBilling{}, errors.New("unexpected call")
	}
	a := f.answers[f.calls]
	f.calls++
	return a.billing, a.err
}

// clock is a settable now, starting 10 days into November (WIB)
type clock struct{ now time.Time }

func (c *clock) Now() time.Time          { return c.now }
func (c *clock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func newUsecase(answers ...fakeAnswer) (*UseCase, *fakeBilling, *clock) {
	billing := &fakeBilling{answers: answers}
	s := New(billing, 12_000_000, time.Hour)
	c := &clock{now: time.Date(2026, 11, 10, 17, 0, 0, 0, time.UTC)}
	s.now = c.Now
	return s, billing, c
}

func TestGet(t *testing.T) {
	s, _, _ := newUsecase(fakeAnswer{billing: domain.HostingBilling{Credit: 6_000_000, HourlyRate: 1_200_000.0 / 720}})

	got, err := s.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Deposit != 18_000_000 || got.MonthlyCost != 1_200_000 {
		t.Errorf("report = %+v", got)
	}
	if got.SafeUntil == nil || *got.SafeUntil != "2028-02-10" {
		t.Errorf("SafeUntil = %v", deref(got.SafeUntil))
	}
}

func TestGetCaches(t *testing.T) {
	s, billing, c := newUsecase(
		fakeAnswer{billing: domain.HostingBilling{Credit: 1_000_000, HourlyRate: 1_200_000.0 / 720}},
		fakeAnswer{billing: domain.HostingBilling{Credit: 2_000_000, HourlyRate: 1_200_000.0 / 720}},
	)

	a, _ := s.Get(context.Background())
	c.Advance(59 * time.Minute)
	b, _ := s.Get(context.Background()) // still cached
	c.Advance(2 * time.Minute)
	fresh, err := s.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if a.Deposit != b.Deposit || billing.calls != 2 {
		t.Errorf("cached deposit = %d, want %d; calls = %d, want 2", b.Deposit, a.Deposit, billing.calls)
	}
	if fresh.Deposit != 14_000_000 {
		t.Errorf("deposit after TTL = %d, want 14000000", fresh.Deposit)
	}
}

func TestGetErrors(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		s := New(nil, 0, 0)
		if _, err := s.Get(context.Background()); codeOf(err) != apperror.CodeUnavailable {
			t.Errorf("error = %v, want UNAVAILABLE", err)
		}
	})

	t.Run("provider down with nothing cached", func(t *testing.T) {
		s, _, _ := newUsecase(fakeAnswer{err: errors.New("idcloudhost: status 502")})
		if _, err := s.Get(context.Background()); codeOf(err) != apperror.CodeUnavailable {
			t.Errorf("error = %v, want UNAVAILABLE", err)
		}
	})

	t.Run("provider down keeps the last report", func(t *testing.T) {
		s, billing, c := newUsecase(
			fakeAnswer{billing: domain.HostingBilling{Credit: 6_000_000, HourlyRate: 1_200_000.0 / 720}},
			fakeAnswer{err: errors.New("timeout")},
		)

		want, _ := s.Get(context.Background())
		c.Advance(2 * time.Hour)
		got, err := s.Get(context.Background())
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got != want {
			t.Errorf("report = %+v, want the last one %+v", got, want)
		}
		// and it is not asked again right away
		c.Advance(time.Minute)
		if _, err := s.Get(context.Background()); err != nil || billing.calls != 2 {
			t.Errorf("Get: %v, calls = %d, want 2", err, billing.calls)
		}
	})
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		return appErr.Code
	}
	return ""
}

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
