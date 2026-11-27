// Package transparency reports where the shared server's money stands, for the public site's
// "server bersama" section
package transparency

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/usecase"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

const defaultCacheTTL = time.Hour

// UseCase asks the provider at most once per cache TTL: the page is public, and the provider's
// API has no published rate limit
type UseCase struct {
	billing      usecase.HostingBillingItf // nil while the provider is not configured
	extraDeposit int64
	ttl          time.Duration
	now          func() time.Time
	location     *time.Location

	mu      sync.Mutex // held while asking, so a burst of visitors asks once
	cached  domain.Transparency
	expires time.Time
}

// New returns the usecase; extraDeposit is rupiah held for the server but not topped up yet. A nil billing makes Get answer UNAVAILABLE; ttl <= 0 means one hour.
func New(billing usecase.HostingBillingItf, extraDeposit int64, ttl time.Duration) *UseCase {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	// months start at midnight WIB, as the team reads them
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		location = time.FixedZone("WIB", 7*60*60)
	}

	return &UseCase{billing: billing, extraDeposit: extraDeposit, ttl: ttl, now: time.Now, location: location}
}

// Get returns the report, from the cache while it is fresh. When the provider cannot be reached,
// the last report is kept for another TTL rather than leaving the page empty.
func (uc *UseCase) Get(ctx context.Context) (domain.Transparency, error) {
	if uc.billing == nil {
		return domain.Transparency{}, apperror.New(apperror.CodeUnavailable, errors.New("hosting billing not configured (IDCLOUDHOST_API_KEY, IDCLOUDHOST_BILLING_ACCOUNT_ID)"), "laporan server belum tersedia")
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	now := uc.now().In(uc.location)
	if now.Before(uc.expires) {
		return uc.cached, nil
	}

	billing, err := uc.billing.Billing(ctx)
	if err != nil {
		if !uc.cached.UpdatedAt.IsZero() {
			log.Printf("transparency: keeping the report of %s: %v", uc.cached.UpdatedAt.Format(time.RFC3339), err)
			uc.expires = now.Add(uc.ttl)
			return uc.cached, nil
		}
		return domain.Transparency{}, apperror.New(apperror.CodeUnavailable, err, "laporan server belum tersedia, silakan coba lagi nanti")
	}

	uc.cached = domain.NewTransparency(billing, uc.extraDeposit, now)
	uc.expires = now.Add(uc.ttl)

	return uc.cached, nil
}
