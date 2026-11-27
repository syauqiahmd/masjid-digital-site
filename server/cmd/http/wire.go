package main

import (
	"log"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/config"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http/controller"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/repository/idcloudhost"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/repository/telegram"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/usecase/inquiry"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/usecase/transparency"
)

// newController wires the repositories into the usecases, and the usecases into the controller
func newController(cfg *config.Schema) *controller.Controller {
	return controller.NewController(newInquiryUsecase(cfg.Inquiry), newTransparencyUsecase(cfg.Transparency))
}

// newInquiryUsecase sends the form to the configured Telegram chat. Without one the form answers
// 503.
func newInquiryUsecase(cfg config.Inquiry) *inquiry.UseCase {
	if cfg.TelegramBotToken == "" || cfg.TelegramChatID == "" {
		log.Print("inquiry form disabled: TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is empty")
		return inquiry.New(nil, cfg.EnvTag)
	}
	return inquiry.New(telegram.New(cfg.TelegramBotToken, cfg.TelegramChatID), cfg.EnvTag)
}

// newTransparencyUsecase reads the shared server's billing account at IDCloudHost. Without an API
// key or billing account the numbers answer 503.
func newTransparencyUsecase(cfg config.Transparency) *transparency.UseCase {
	if cfg.IDCloudHostAPIKey == "" || cfg.BillingAccountID == 0 {
		log.Print("transparency numbers disabled: IDCLOUDHOST_API_KEY or IDCLOUDHOST_BILLING_ACCOUNT_ID is empty")
		return transparency.New(nil, cfg.ExtraDeposit, cfg.CacheTTL)
	}
	return transparency.New(idcloudhost.New(cfg.IDCloudHostAPIKey, cfg.BillingAccountID, cfg.VMNames), cfg.ExtraDeposit, cfg.CacheTTL)
}
