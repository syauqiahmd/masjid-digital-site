package domain

import (
	"math"
	"time"
)

// averageMonth is how long the deposit's months are when turned into a date
const averageMonth = time.Duration(365.2425 / 12 * 24 * 3600 * 1e9) // 730h29m06s

// HostingBilling is what the hosting provider reports for the shared server's billing account,
// in rupiah
type HostingBilling struct {
	Credit     float64 // prepaid credit left after this month's charges so far
	HourlyRate float64 // what the shared server's resources cost per hour now, before VAT
	VATPercent float64 // e.g. 11 for 11%
}

// MonthlyUsage is a full month of the current hourly rate, before VAT: what the credit goes down
// by. The month is now's month, so it is 720 or 744 hours as the provider bills it.
func (b HostingBilling) MonthlyUsage(now time.Time) float64 {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	month := start.AddDate(0, 1, 0).Sub(start)

	return b.HourlyRate * month.Hours()
}

// MonthlyCost is MonthlyUsage with VAT: what the month costs in money, since the provider charges
// VAT when the credit is bought rather than when it is used.
func (b HostingBilling) MonthlyCost(now time.Time) float64 {
	return b.MonthlyUsage(now) * (1 + b.VATPercent/100)
}

// Transparency is the "server bersama" report: where the shared server's money stands. Amounts
// are whole rupiah.
type Transparency struct {
	Deposit     int64   `json:"deposit"`      // credit at the provider plus what is not topped up yet
	MonthlyCost int64   `json:"monthly_cost"` // a month at the current rate, VAT included
	SafeUntil   *string `json:"safe_until"`   // the day the deposit runs out at this cost, e.g. "2028-03-14"; nil when nothing is charged
	// UpdatedAt is when the provider was asked; the answer is cached
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTransparency adds extraDeposit (rupiah held for the server but not topped up yet) to the
// provider's credit and works out how long it lasts
func NewTransparency(billing HostingBilling, extraDeposit int64, now time.Time) Transparency {
	deposit := math.Round(billing.Credit) + float64(extraDeposit)
	usage := billing.MonthlyUsage(now) // the credit is used up before VAT

	report := Transparency{
		Deposit:     int64(deposit),
		MonthlyCost: int64(math.Round(billing.MonthlyCost(now))),
		UpdatedAt:   now,
	}
	if usage > 0 {
		months := max(deposit, 0) / usage
		// past a century the date stops meaning anything, and Duration would overflow
		months = min(months, 1200)
		safeUntil := now.Add(time.Duration(months * float64(averageMonth))).Format(time.DateOnly)
		report.SafeUntil = &safeUntil
	}

	return report
}
