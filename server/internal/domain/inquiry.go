package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

const (
	maxInquiryLineLength  = 100  // names, city, source, domain
	maxInquiryNotesLength = 1000 // fits a Telegram message (4096) with room for the rest

	// MinDonation is the least a Donatur pays per year, in rupiah: enough for the domain, the
	// rest goes to the shared server's deposit
	MinDonation = 1_000_000
	maxDonation = 1_000_000_000_000 // past this it is a typo
)

// InquiryPlan is the plan a mosque asks for
type InquiryPlan string

const (
	InquiryPlanGratis  InquiryPlan = "gratis"  // a subdomain of masjidyuk.com on the shared server
	InquiryPlanDonatur InquiryPlan = "donatur" // its own domain, paid for by a yearly donation
)

// Label is the plan as the team calls it
func (p InquiryPlan) Label() string {
	switch p {
	case InquiryPlanGratis:
		return "Gratis"
	case InquiryPlanDonatur:
		return "Donatur"
	}
	return string(p)
}

// InquiryRequest is what the Daftar form sends. Nothing of it is stored: the usecase forwards it
// to the team's Telegram chat.
type InquiryRequest struct {
	MosqueName string      `json:"mosque_name"`
	Plan       InquiryPlan `json:"plan"`
	Subdomain  string      `json:"subdomain"` // Gratis, optional: the part before .masjidyuk.com
	Domain     string      `json:"domain"`    // Donatur, optional: the domain they want
	Donation   int64       `json:"donation"`  // Donatur: rupiah per year, at least MinDonation
	FullName   string      `json:"full_name"`
	Phone      string      `json:"phone"`
	Domicile   string      `json:"domicile"`
	Source     string      `json:"source"` // optional: where they heard of Masjid Digital
	Notes      string      `json:"notes"`  // optional
	Consent    bool        `json:"consent"`
	// Website is a honeypot: the form hides it from people, so only bots fill it in
	Website string `json:"website"`
}

// Inquiry is a validated InquiryRequest, ready to forward. Fields of the other plan are empty.
type Inquiry struct {
	MosqueName string
	Plan       InquiryPlan
	Subdomain  string // lower case, e.g. alikhlas
	Domain     string // lower case, e.g. alikhlas.id
	Donation   int64
	FullName   string
	Phone      string // E.164, e.g. +6281234567890
	Domicile   string
	Source     string
	Notes      string
}

// IsBot tells whether the honeypot field was filled in
func (r InquiryRequest) IsBot() bool {
	return strings.TrimSpace(r.Website) != ""
}

// Validate trims and checks every field, normalizing the phone number to E.164. One-line fields
// have their whitespace collapsed, so a name cannot add lines to the message.
func (r InquiryRequest) Validate() (Inquiry, error) {
	inquiry := Inquiry{
		MosqueName: oneLine(r.MosqueName),
		Plan:       InquiryPlan(strings.ToLower(strings.TrimSpace(string(r.Plan)))),
		FullName:   oneLine(r.FullName),
		Domicile:   oneLine(r.Domicile),
		Source:     oneLine(r.Source),
		Notes:      strings.TrimSpace(r.Notes),
	}

	required := []struct{ value, label string }{
		{inquiry.MosqueName, "nama masjid"},
		{string(inquiry.Plan), "paket"},
		{inquiry.FullName, "nama pengurus"},
		{r.Phone, "nomor WhatsApp"},
		{inquiry.Domicile, "kota"},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return Inquiry{}, invalid(field.label + " wajib diisi")
		}
	}

	limits := []struct {
		value, label string
		max          int
	}{
		{inquiry.MosqueName, "nama masjid", maxInquiryLineLength},
		{inquiry.FullName, "nama pengurus", maxInquiryLineLength},
		{inquiry.Domicile, "kota", maxInquiryLineLength},
		{inquiry.Source, "informasi lainnya", maxInquiryLineLength},
		{inquiry.Notes, "catatan", maxInquiryNotesLength},
	}
	for _, field := range limits {
		if utf8.RuneCountInString(field.value) > field.max {
			return Inquiry{}, invalid(fmt.Sprintf("%s maksimal %d karakter", field.label, field.max))
		}
	}

	switch inquiry.Plan {
	case InquiryPlanGratis:
		inquiry.Subdomain = strings.ToLower(strings.TrimSpace(r.Subdomain))
		if inquiry.Subdomain != "" && !subdomainPattern.MatchString(inquiry.Subdomain) {
			return Inquiry{}, invalid("alamat hanya boleh huruf kecil, angka, dan tanda hubung, 3 sampai 63 karakter")
		}
	case InquiryPlanDonatur:
		inquiry.Domain = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(r.Domain)), ".")
		if inquiry.Domain != "" && (len(inquiry.Domain) > 253 || !domainPattern.MatchString(inquiry.Domain)) {
			return Inquiry{}, invalid("domain tidak valid, contoh: masjidkamu.id")
		}
		if r.Donation < MinDonation || r.Donation > maxDonation {
			return Inquiry{}, invalid("donasi minimal Rp1.000.000 per tahun untuk paket Donatur")
		}
		inquiry.Donation = r.Donation
	default:
		return Inquiry{}, invalid("paket tidak dikenal")
	}

	phone, ok := normalizePhone(r.Phone)
	if !ok {
		return Inquiry{}, invalid("nomor WhatsApp tidak valid")
	}
	inquiry.Phone = phone

	if !r.Consent {
		return Inquiry{}, invalid("centang persetujuan untuk dihubungi")
	}

	return inquiry, nil
}

var (
	// subdomainPattern is one DNS label of at least 3 characters, e.g. alikhlas or al-ikhlas
	subdomainPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,61}[a-z0-9])$`)
	// domainPattern is two or more DNS labels with a letters-only top level, e.g. alikhlas.or.id
	domainPattern = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)
)

// WhatsAppURL opens a chat with the sender
func (i Inquiry) WhatsAppURL() string {
	return "https://wa.me/" + strings.TrimPrefix(i.Phone, "+")
}

// oneLine trims s and collapses every run of whitespace, line breaks included, to one space
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

var (
	// phoneSeparators are stripped from a phone number before validating it
	phoneSeparators = strings.NewReplacer(" ", "", "-", "", ".", "", "(", "", ")", "")
	// e164 is a + followed by a country code and up to 15 digits in total
	e164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	// indonesianMobile is +62 8xx with 9 to 12 digits after the country code
	indonesianMobile = regexp.MustCompile(`^\+628\d{8,11}$`)
)

// normalizePhone converts a phone number to E.164. Numbers without a + are read as Indonesian:
// 0812-3456-7890 and 6281234567890 both become +6281234567890.
func normalizePhone(phone string) (string, bool) {
	p := phoneSeparators.Replace(strings.TrimSpace(phone))

	switch {
	case strings.HasPrefix(p, "+"):
	case strings.HasPrefix(p, "0"):
		p = "+62" + p[1:]
	case strings.HasPrefix(p, "62"):
		p = "+" + p
	default:
		return "", false
	}

	if !e164.MatchString(p) {
		return "", false
	}
	if strings.HasPrefix(p, "+62") && !indonesianMobile.MatchString(p) {
		return "", false
	}

	return p, true
}

// invalid is an error the sender has to fix, with message shown to them as is
func invalid(message string) error {
	return apperror.New(apperror.CodeInvalidInput, errors.New(message), message)
}
