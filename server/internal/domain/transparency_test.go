package domain

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestHostingBillingMonthlyCost(t *testing.T) {
	wib := time.FixedZone("WIB", 7*60*60)

	tests := []struct {
		name    string
		billing HostingBilling
		now     time.Time
		want    float64
	}{
		{"30-day month", HostingBilling{HourlyRate: 1_000}, time.Date(2026, 11, 27, 15, 0, 0, 0, wib), 720_000},
		{"31-day month", HostingBilling{HourlyRate: 1_000}, time.Date(2026, 12, 1, 0, 0, 0, 0, wib), 744_000},
		{"adds VAT", HostingBilling{HourlyRate: 1_000, VATPercent: 11}, time.Date(2026, 11, 1, 0, 0, 0, 0, wib), 799_200},
		{"nothing running", HostingBilling{Credit: 5_000_000}, time.Date(2026, 11, 11, 0, 0, 0, 0, wib), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.billing.MonthlyCost(tt.now); math.Abs(got-tt.want) > 0.001 {
				t.Errorf("MonthlyCost = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTransparency(t *testing.T) {
	wib := time.FixedZone("WIB", 7*60*60)
	now := time.Date(2026, 11, 11, 0, 0, 0, 0, wib)

	t.Run("deposit lasts deposit / monthly cost months", func(t *testing.T) {
		// 1.2 jt a month; 6 jt at the provider + 12 jt in the bank = 15 months
		billing := HostingBilling{Credit: 6_000_000.4, HourlyRate: 1_200_000.0 / 720}
		got := NewTransparency(billing, 12_000_000, now)

		if got.Deposit != 18_000_000 || got.MonthlyCost != 1_200_000 {
			t.Errorf("deposit, monthly = %d, %d; want 18000000, 1200000", got.Deposit, got.MonthlyCost)
		}
		if got.SafeUntil == nil || *got.SafeUntil != "2028-02-10" {
			t.Errorf("SafeUntil = %v, want 2028-02-10", deref(got.SafeUntil))
		}
		if !got.UpdatedAt.Equal(now) {
			t.Errorf("report = %+v", got)
		}
	})

	t.Run("VAT shows in the cost but not in how long the credit lasts", func(t *testing.T) {
		// VAT is paid when the credit is bought, so 1.2 jt of credit lasts a month of 1.2 jt usage
		got := NewTransparency(HostingBilling{Credit: 1_200_000, HourlyRate: 1_200_000.0 / 720, VATPercent: 11}, 0, now)
		if got.MonthlyCost != 1_332_000 {
			t.Errorf("MonthlyCost = %d, want 1332000", got.MonthlyCost)
		}
		if got.SafeUntil == nil || *got.SafeUntil != "2026-12-11" {
			t.Errorf("SafeUntil = %v, want 2026-12-11", deref(got.SafeUntil))
		}
	})

	t.Run("nothing charged has no end date", func(t *testing.T) {
		got := NewTransparency(HostingBilling{Credit: 1_000_000}, 0, now)
		if got.SafeUntil != nil {
			t.Errorf("SafeUntil = %v, want nil", *got.SafeUntil)
		}
	})

	t.Run("an empty or negative deposit ends today", func(t *testing.T) {
		got := NewTransparency(HostingBilling{Credit: -50_000, HourlyRate: 1_200_000.0 / 720}, 0, now)
		if got.SafeUntil == nil || *got.SafeUntil != "2026-11-11" {
			t.Errorf("SafeUntil = %v, want 2026-11-11", deref(got.SafeUntil))
		}
	})

	t.Run("JSON", func(t *testing.T) {
		got, err := json.Marshal(NewTransparency(HostingBilling{Credit: 1_200_000, HourlyRate: 1_200_000.0 / 720}, 0, now))
		if err != nil {
			t.Fatal(err)
		}
		want := `{"deposit":1200000,"monthly_cost":1200000,"safe_until":"2026-12-11","updated_at":"2026-11-11T00:00:00+07:00"}`
		if string(got) != want {
			t.Errorf("JSON = %s\nwant   %s", got, want)
		}
	})
}

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
