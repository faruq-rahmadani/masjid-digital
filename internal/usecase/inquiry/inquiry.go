// Package inquiry forwards the public site's inquiry form to the team. Nothing is stored.
package inquiry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// UseCase validates an inquiry and sends it to the team's chat
type UseCase struct {
	notifier usecase.InquiryNotifierItf // nil while the chat is not configured
	envTag   string
	now      func() time.Time
	location *time.Location
}

// New returns the usecase. A nil notifier makes Submit answer UNAVAILABLE. A non-empty envTag
// (e.g. "staging") starts each message with [STAGING], so test submissions stand out.
func New(notifier usecase.InquiryNotifierItf, envTag string) *UseCase {
	// the team reads the time in WIB; distroless images ship tzdata
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		location = time.FixedZone("WIB", 7*60*60)
	}

	return &UseCase{notifier: notifier, envTag: strings.TrimSpace(envTag), now: time.Now, location: location}
}

// Submit validates req and sends it. A request with the honeypot filled in is dropped
// silently, so a bot sees the same answer as a person.
func (uc *UseCase) Submit(ctx context.Context, req domain.InquiryRequest) error {
	if uc.notifier == nil {
		return apperror.New(apperror.CodeUnavailable, errors.New("inquiry notifier not configured (app.inquiry)"), "formulir belum bisa dipakai, silakan coba lagi nanti")
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

// message is the plain text the team reads
func (uc *UseCase) message(inquiry domain.Inquiry) string {
	var b strings.Builder

	if uc.envTag != "" {
		fmt.Fprintf(&b, "[%s] ", strings.ToUpper(uc.envTag))
	}
	b.WriteString("📩 Inquiry baru dari halaman publik\n\n")

	fmt.Fprintf(&b, "Nama: %s\n", inquiry.FullName)
	fmt.Fprintf(&b, "Ponsel: %s\n", inquiry.Phone)
	fmt.Fprintf(&b, "WhatsApp: %s\n", inquiry.WhatsAppURL())
	fmt.Fprintf(&b, "Domisili: %s\n", inquiry.Domicile)
	if inquiry.Source != "" {
		fmt.Fprintf(&b, "Tahu dari: %s\n", inquiry.Source)
	}
	fmt.Fprintf(&b, "\nAlasan:\n%s\n\n", inquiry.Reason)
	fmt.Fprintf(&b, "Dikirim %s WIB", uc.now().In(uc.location).Format("2 Jan 2006 15:04"))

	return b.String()
}
