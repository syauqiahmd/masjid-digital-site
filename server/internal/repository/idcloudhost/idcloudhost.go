// Package idcloudhost reads one billing account from the IDCloudHost API, see
// https://api.idcloudhost.com/. Only GET endpoints are called, but the API key itself can do
// everything the account can, so it must stay on the server.
package idcloudhost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
)

const (
	defaultBaseURL = "https://api.idcloudhost.com"
	requestTimeout = 10 * time.Second
	maxBodySize    = 4 << 20
)

// Client reads one billing account
type Client struct {
	apiKey           string
	billingAccountID string
	vmNames          map[string]bool // empty: every VM
	baseURL          string
	client           *http.Client
}

// New returns a Client for the billing account with that ID. With vmNames, only those VMs'
// resources count toward the hourly rate, for an account that also pays for other projects.
func New(apiKey string, billingAccountID int64, vmNames []string) *Client {
	names := map[string]bool{}
	for _, name := range vmNames {
		names[name] = true
	}
	return &Client{
		apiKey:           apiKey,
		billingAccountID: strconv.FormatInt(billingAccountID, 10),
		vmNames:          names,
		baseURL:          defaultBaseURL,
		client:           &http.Client{Timeout: requestTimeout},
	}
}

// billingAccount is the part of GET /v1/payment/billing_account that is used. The rest holds
// the owner's address and card, which never leave this package.
type billingAccount struct {
	VATPercentage float64 `json:"vat_percentage"`
	RunningTotals struct {
		// Ongoing is the credit left after this month's charges so far; credit_available is the
		// credit before them
		Ongoing float64 `json:"ongoing"`
	} `json:"running_totals"`
}

// usageItem is one resource of GET /v1/charging/usage, this month
type usageItem struct {
	UUID           string  `json:"uuid"`       // the resource: a VM, a disk, one backup replica
	Price          float64 `json:"price"`      // per price_unit, before VAT
	PriceUnit      string  `json:"price_unit"` // "h" for resources; "pc" for one-off products
	Configurations []struct {
		VMName *string `json:"vm_name"`
	} `json:"configurations"`
}

func (item usageItem) vmName() string {
	for _, c := range item.Configurations {
		if c.VMName != nil && *c.VMName != "" {
			return *c.VMName
		}
	}
	return ""
}

// Billing returns the account's credit left and the hourly rate of the VMs' resources billed this
// month: the VM itself, its disk and each backup replica. One-off products (a carried-over unpaid
// report, say) have no VM and do not count.
func (c *Client) Billing(ctx context.Context) (domain.HostingBilling, error) {
	var account billingAccount
	if err := c.get(ctx, "/v1/payment/billing_account", &account); err != nil {
		return domain.HostingBilling{}, err
	}

	var usage []usageItem
	if err := c.get(ctx, "/v1/charging/usage", &usage); err != nil {
		return domain.HostingBilling{}, err
	}

	rates := map[string]float64{}
	for _, item := range usage {
		vm := item.vmName()
		if item.PriceUnit != "h" || vm == "" || (len(c.vmNames) > 0 && !c.vmNames[vm]) {
			continue
		}
		rates[item.UUID] = item.Price
	}

	billing := domain.HostingBilling{Credit: account.RunningTotals.Ongoing, VATPercent: account.VATPercentage}
	for _, rate := range rates {
		billing.HourlyRate += rate
	}

	return billing, nil
}

// get decodes the JSON answer of path for the billing account into out. Its errors never contain
// the API key.
func (c *Client) get(ctx context.Context, path string, out any) error {
	query := url.Values{"billing_account_id": {c.billingAccountID}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("idcloudhost: %s: %w", path, err)
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			err = urlErr.Err
		}
		return fmt.Errorf("idcloudhost: %s: %w", path, err)
	}
	defer res.Body.Close()

	body := io.LimitReader(res.Body, maxBodySize)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("idcloudhost: %s: status %d", path, res.StatusCode)
	}
	if err := json.NewDecoder(body).Decode(out); err != nil {
		return fmt.Errorf("idcloudhost: %s: decode: %w", path, err)
	}

	return nil
}
