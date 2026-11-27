// Package usecase holds the application's rules; its subpackages depend on the interfaces here,
// which the repositories implement.
package usecase

import (
	"context"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
)

// InquiryNotifierItf delivers an inquiry from the public site to the team, e.g. a Telegram chat
type InquiryNotifierItf interface {
	Send(ctx context.Context, text string) error
}

// HostingBillingItf reads the shared server's billing account at the hosting provider
type HostingBillingItf interface {
	Billing(ctx context.Context) (domain.HostingBilling, error)
}
