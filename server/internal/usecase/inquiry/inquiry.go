// Package inquiry forwards the public site's Daftar form to the team. Nothing is stored.
package inquiry

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/domain"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/usecase"
	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

// UseCase validates an inquiry and sends it to the team's chat
type UseCase struct {
	notifier usecase.InquiryNotifierItf // nil while the chat is not configured
	envTag   string
	now      func() time.Time
	location *time.Location
}

// New returns the usecase. A nil notifier makes Submit answer UNAVAILABLE. A non-empty
// envTag (e.g. "staging") starts each message with [STAGING], so test submissions stand out.
func New(notifier usecase.InquiryNotifierItf, envTag string) *UseCase {
	// the team reads the time in WIB; main embeds tzdata
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		location = time.FixedZone("WIB", 7*60*60)
	}

	return &UseCase{notifier: notifier, envTag: strings.TrimSpace(envTag), now: time.Now, location: location}
}

// Submit validates req and sends it. A request with the honeypot filled in is dropped silently,
// so a bot sees the same answer as a person.
func (uc *UseCase) Submit(ctx context.Context, req domain.InquiryRequest) error {
	if uc.notifier == nil {
		return apperror.New(apperror.CodeUnavailable, errors.New("inquiry notifier not configured (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID)"), "formulir belum bisa dipakai, silakan coba lagi nanti")
	}

	inquiry, err := req.Validate()
	if err != nil {
		return err
	}

	if req.IsBot() {
		log.Printf("inquiry dropped: honeypot filled in")
		return nil
	}

	if err := uc.notifier.Send(ctx, uc.message(inquiry)); err != nil {
		return apperror.New(apperror.CodeUnavailable, err, "pesan belum terkirim, silakan coba lagi beberapa saat lagi")
	}

	return nil
}

// message is what the team reads, in Telegram HTML: the mosque and its plan first, then who to
// contact. Every value a visitor typed is escaped, so it shows as typed and never as markup.
func (uc *UseCase) message(inquiry domain.Inquiry) string {
	esc := html.EscapeString
	var b strings.Builder

	if uc.envTag != "" {
		fmt.Fprintf(&b, "<code>%s</code> ", esc(strings.ToUpper(uc.envTag)))
	}
	b.WriteString("📩 <b>Pendaftaran baru</b>\n\n")

	fmt.Fprintf(&b, "🕌 <b>%s</b>\n", esc(inquiry.MosqueName))
	fmt.Fprintf(&b, "Paket: <b>%s</b>\n", esc(inquiry.Plan.Label()))
	switch inquiry.Plan {
	case domain.InquiryPlanGratis:
		fmt.Fprintf(&b, "Alamat: %s\n", orUndecided(inquiry.Subdomain, ".masjidyuk.com"))
	case domain.InquiryPlanDonatur:
		fmt.Fprintf(&b, "Domain: %s\n", orUndecided(inquiry.Domain, ""))
		fmt.Fprintf(&b, "Donasi: <b>%s</b> / tahun\n", rupiah(inquiry.Donation))
	}

	fmt.Fprintf(&b, "\n👤 <b>%s</b>\n", esc(inquiry.FullName))
	fmt.Fprintf(&b, "WhatsApp: <a href=\"%s\">%s</a>\n", esc(inquiry.WhatsAppURL()), esc(inquiry.Phone))
	fmt.Fprintf(&b, "Kota: %s\n", esc(inquiry.Domicile))
	if inquiry.Source != "" {
		fmt.Fprintf(&b, "Tahu dari: %s\n", esc(inquiry.Source))
	}
	if inquiry.Notes != "" {
		fmt.Fprintf(&b, "\n📝 <b>Catatan</b>\n<blockquote>%s</blockquote>\n", esc(inquiry.Notes))
	}

	fmt.Fprintf(&b, "\n<i>Dikirim %s WIB</i>", uc.now().In(uc.location).Format("2 Jan 2006 15:04"))

	return b.String()
}

// orUndecided is value with its suffix in code style (easy to copy), or a note that the team
// still has to agree on it
func orUndecided(value, suffix string) string {
	if value == "" {
		return "<i>belum ditentukan</i>"
	}
	return "<code>" + html.EscapeString(value+suffix) + "</code>"
}

// rupiah writes 1500000 as Rp1.500.000
func rupiah(amount int64) string {
	digits := strconv.FormatInt(amount, 10)
	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	return "Rp" + b.String()
}
