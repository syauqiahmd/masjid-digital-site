package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("TELEGRAM_BOT_TOKEN", " 123:abc ")
	t.Setenv("IDCLOUDHOST_BILLING_ACCOUNT_ID", "42")
	t.Setenv("TRANSPARENCY_EXTRA_DEPOSIT", "12.000.000")
	t.Setenv("TRANSPARENCY_CACHE_TTL", "30m")
	t.Setenv("TRANSPARENCY_VM_NAMES", " masjidyuk, cadangan ,")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPServer.Port != "9000" || cfg.Inquiry.TelegramBotToken != "123:abc" {
		t.Errorf("config = %+v", cfg)
	}
	tr := cfg.Transparency
	if tr.BillingAccountID != 42 || tr.ExtraDeposit != 12_000_000 || tr.CacheTTL != 30*time.Minute {
		t.Errorf("transparency = %+v", tr)
	}
	if len(tr.VMNames) != 2 || tr.VMNames[0] != "masjidyuk" || tr.VMNames[1] != "cadangan" {
		t.Errorf("VMNames = %q", tr.VMNames)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPServer.Port != "8080" || cfg.Transparency.CacheTTL != time.Hour || cfg.Transparency.BillingAccountID != 0 {
		t.Errorf("config = %+v", cfg)
	}
}

func TestLoadRejectsBadNumbers(t *testing.T) {
	t.Setenv("TRANSPARENCY_EXTRA_DEPOSIT", "dua juta")
	if _, err := Load(); err == nil {
		t.Error("Load error = nil, want one")
	}
}
