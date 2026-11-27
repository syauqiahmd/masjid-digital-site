package idcloudhost

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testKey = "secret-api-key"

// the shapes of the examples in the API docs, cut down
const (
	// as a real account answers: ongoing is credit_available less this month's charges
	billingAccountJSON = `{"id":6,"credit_amount":200000.0,"email":"user@example.com","primary_card":{"last4":"1111"},
		"running_totals":{"charging_subtotal":49010.39,"credit_amount":-49010.39,"credit_available":200000.0,"ongoing":150989.61,"subtotal":49010.39,"total":0.0,"vat_tax":0.0},
		"unpaid_amount":0,"vat_percentage":11.0}`
	// as a real account answers: the shared VM, its disk and its two backup replicas; another
	// project's VM; a one-off product
	usageJSON = `[
		{"uuid":"vm-1","cost":29589.12,"price":102.74,"price_unit":"h","hours":288.0,"uptime_types":["VM_COMPUTE"],"configurations":[{"vm_name":"masjidyuk","cpus":2,"ram_MB":2048,"disk_size_GB":null}]},
		{"uuid":"disk-1","cost":4947.68,"price":17.12,"price_unit":"h","hours":289.0,"uptime_types":["BLOCK_STORAGE"],"configurations":[{"vm_name":"masjidyuk"}]},
		{"uuid":"backup-1","cost":196.88,"price":0.856,"price_unit":"h","hours":230.0,"uptime_types":["REPLICA_BACKUP"],"configurations":[{"vm_name":"masjidyuk"}]},
		{"uuid":"backup-2","cost":53.07,"price":0.856,"price_unit":"h","hours":62.0,"uptime_types":["REPLICA_BACKUP"],"configurations":[{"vm_name":"masjidyuk"}]},
		{"uuid":"vm-2","cost":308.22,"price":102.74,"price_unit":"h","hours":3.0,"uptime_types":["VM_COMPUTE"],"configurations":[{"vm_name":"other-project"}]},
		{"uuid":"report-1","cost":9269.07,"price":9269.07,"price_unit":"pc","hours":1.0,"uptime_types":["CUSTOM_PRODUCT"],"configurations":[{"vm_name":null}]}
	]`
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c := New(testKey, 6, []string{"masjidyuk"})
	c.baseURL = server.URL
	return c
}

func TestBilling(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("apikey") != testKey || r.URL.Query().Get("billing_account_id") != "6" {
			t.Errorf("request = %s %s, apikey %q", r.Method, r.URL, r.Header.Get("apikey"))
		}
		switch r.URL.Path {
		case "/v1/payment/billing_account":
			w.Write([]byte(billingAccountJSON))
		case "/v1/charging/usage":
			w.Write([]byte(usageJSON))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
		}
	})

	got, err := c.Billing(context.Background())
	if err != nil {
		t.Fatalf("Billing: %v", err)
	}
	// 102.74 + 17.12 + 2 × 0.856, without the other project and the one-off product
	if got.Credit != 150989.61 || got.VATPercent != 11 || math.Abs(got.HourlyRate-121.572) > 1e-9 {
		t.Errorf("Billing = %+v, want credit 150989.61, rate 121.572, VAT 11", got)
	}
}

func TestBillingEveryVM(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/charging/usage" {
			w.Write([]byte(usageJSON))
			return
		}
		w.Write([]byte(billingAccountJSON))
	})
	c.vmNames = map[string]bool{}

	got, err := c.Billing(context.Background())
	if err != nil {
		t.Fatalf("Billing: %v", err)
	}
	if math.Abs(got.HourlyRate-224.312) > 1e-9 {
		t.Errorf("HourlyRate = %v, want 224.312 with the other project's VM", got.HourlyRate)
	}
}

func TestBillingErrors(t *testing.T) {
	t.Run("rejected key", func(t *testing.T) {
		c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"Invalid authentication credentials"}`))
		})
		_, err := c.Billing(context.Background())
		if err == nil || !strings.Contains(err.Error(), "status 401") {
			t.Errorf("Billing error = %v, want status 401", err)
		}
	})

	t.Run("usage is not a list", func(t *testing.T) {
		c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/charging/usage" {
				w.Write([]byte(`{"error":"x"}`))
				return
			}
			w.Write([]byte(billingAccountJSON))
		})
		if _, err := c.Billing(context.Background()); err == nil || !strings.Contains(err.Error(), "decode") {
			t.Errorf("Billing error = %v, want a decode error", err)
		}
	})

	t.Run("unreachable", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		server.Close()

		c := New(testKey, 6, nil)
		c.baseURL = server.URL
		_, err := c.Billing(context.Background())
		if err == nil {
			t.Fatal("Billing error = nil, want one")
		}
		if strings.Contains(err.Error(), testKey) {
			t.Errorf("error leaks the API key: %v", err)
		}
	})
}
