// Package config reads the server's settings from the environment, see .env.example.
//
// Development: copy server/.env.example to server/.env and run from server/; Load reads it.
// Production: docker compose passes the repo root's .env to the api container as its environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Schema struct {
	HTTPServer   HTTPServer
	Inquiry      Inquiry
	Transparency Transparency
}

type HTTPServer struct {
	Port string // PORT, default 8080
}

// Inquiry sends the Daftar form to a Telegram chat; it is not stored. The form answers 503 while
// the bot token or chat ID is empty.
type Inquiry struct {
	TelegramBotToken string // TELEGRAM_BOT_TOKEN, from @BotFather
	TelegramChatID   string // TELEGRAM_CHAT_ID: a user, or a group with the bot in it (negative ID)
	EnvTag           string // INQUIRY_ENV_TAG, e.g. staging, shown as [STAGING]; empty in production
}

// Transparency feeds the "server bersama" numbers: the shared server's credit and running cost
// from IDCloudHost, plus the deposit not topped up yet. The endpoint answers 503 while the API key or
// billing account is empty.
type Transparency struct {
	IDCloudHostAPIKey string        // IDCLOUDHOST_API_KEY; it can also change the account, so keep it secret
	BillingAccountID  int64         // IDCLOUDHOST_BILLING_ACCOUNT_ID
	VMNames           []string      // TRANSPARENCY_VM_NAMES, comma separated: the shared server's VMs; empty means every VM
	ExtraDeposit      int64         // TRANSPARENCY_EXTRA_DEPOSIT: rupiah not topped up to IDCloudHost yet
	CacheTTL          time.Duration // TRANSPARENCY_CACHE_TTL, default 1h
}

// Load reads the environment, after filling what is unset from .env in the working directory.
// Whole numbers may group digits with dots or underscores, e.g. 12.000.000.
func Load() (*Schema, error) {
	if err := loadDotEnv(".env"); err != nil {
		return nil, err
	}

	r := reader{}
	cfg := &Schema{
		HTTPServer: HTTPServer{Port: r.str("PORT", "8080")},
		Inquiry: Inquiry{
			TelegramBotToken: r.str("TELEGRAM_BOT_TOKEN", ""),
			TelegramChatID:   r.str("TELEGRAM_CHAT_ID", ""),
			EnvTag:           r.str("INQUIRY_ENV_TAG", ""),
		},
		Transparency: Transparency{
			IDCloudHostAPIKey: r.str("IDCLOUDHOST_API_KEY", ""),
			BillingAccountID:  r.int("IDCLOUDHOST_BILLING_ACCOUNT_ID"),
			VMNames:           r.list("TRANSPARENCY_VM_NAMES"),
			ExtraDeposit:      r.int("TRANSPARENCY_EXTRA_DEPOSIT"),
			CacheTTL:          r.duration("TRANSPARENCY_CACHE_TTL", time.Hour),
		},
	}
	if r.err != nil {
		return nil, r.err
	}
	return cfg, nil
}

// reader keeps the first bad value, so Load reports it instead of starting half configured
type reader struct {
	err error
}

func (r *reader) str(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}

func (r *reader) list(name string) []string {
	var items []string
	for item := range strings.SplitSeq(r.str(name, ""), ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func (r *reader) int(name string) int64 {
	v := strings.NewReplacer(".", "", "_", "").Replace(r.str(name, "0"))
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil && r.err == nil {
		r.err = fmt.Errorf("%s: %q is not a whole number", name, os.Getenv(name))
	}
	return n
}

func (r *reader) duration(name string, fallback time.Duration) time.Duration {
	v := r.str(name, "")
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil && r.err == nil {
		r.err = fmt.Errorf("%s: %w", name, err)
	}
	return d
}
