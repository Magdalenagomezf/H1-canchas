package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/email"
)

// fakeNotificationRepo tracks claimed state like the real DB column.
type fakeNotificationRepo struct {
	detail     *domain.BookingDetail
	getErr     error
	claimed    bool
	claimErr   error
	claimCalls int
	releases   int
}

func (f *fakeNotificationRepo) GetByIDWithDetails(_ context.Context, _ int64) (*domain.BookingDetail, error) {
	return f.detail, f.getErr
}

func (f *fakeNotificationRepo) ClaimConfirmationEmail(_ context.Context, _ int64) (bool, error) {
	f.claimCalls++
	if f.claimErr != nil {
		return false, f.claimErr
	}
	if f.claimed {
		return false, nil
	}
	f.claimed = true
	return true, nil
}

func (f *fakeNotificationRepo) ReleaseConfirmationEmail(_ context.Context, _ int64) error {
	f.releases++
	f.claimed = false
	return nil
}

type sentEmail struct{ to, subject, html, text string }

type fakeEmailSender struct {
	sent  []sentEmail
	err   error
	calls int
	keys  []string
}

func (f *fakeEmailSender) Send(ctx context.Context, to, subject, html, text string) error {
	f.calls++
	f.keys = append(f.keys, email.IdempotencyKeyFromContext(ctx))
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, sentEmail{to, subject, html, text})
	return nil
}

func confirmedDetail() *domain.BookingDetail {
	email := "ana@example.com"
	name := "Ana"
	start, end := "18:00:00", "19:00:00"
	return &domain.BookingDetail{
		ID: 7,
		// Sábado 4 de octubre de 2025, como lo escanea Postgres (DATE -> 00:00 UTC).
		BookingDate:       time.Date(2025, time.October, 4, 0, 0, 0, 0, time.UTC),
		SpaceName:         "Cancha 1",
		SlotLabel:         "18 a 19",
		SlotStartTime:     &start,
		SlotEndTime:       &end,
		DepositAmount:     1500,
		DepositStatus:     domain.PaymentStatusPaid,
		BalanceAmount:     12500.5,
		BalanceStatus:     domain.PaymentStatusUnpaid,
		CustomerUserName:  &name,
		CustomerUserEmail: &email,
	}
}

func newNotificationSvc(repo *fakeNotificationRepo, sender *fakeEmailSender) *NotificationService {
	return NewNotificationService(repo, sender, "https://app.example.com/")
}

func TestNotificationService_SendBookingConfirmation_SendsOnceWithContent(t *testing.T) {
	repo := &fakeNotificationRepo{detail: confirmedDetail()}
	sender := &fakeEmailSender{}
	svc := newNotificationSvc(repo, sender)

	if err := svc.SendBookingConfirmation(context.Background(), 7); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := svc.SendBookingConfirmation(context.Background(), 7); err != nil {
		t.Fatalf("second call: %v", err)
	}

	if len(sender.sent) != 1 {
		t.Fatalf("got %d emails, want exactly 1", len(sender.sent))
	}
	m := sender.sent[0]
	if m.to != "ana@example.com" {
		t.Errorf("got recipient %q", m.to)
	}
	wantSubject := "Reserva confirmada — Cancha 1, sábado 4 de octubre"
	if m.subject != wantSubject {
		t.Errorf("got subject %q, want %q", m.subject, wantSubject)
	}
	for _, body := range []struct{ name, content string }{{"html", m.html}, {"text", m.text}} {
		for _, want := range []string{
			"Cancha 1", "sábado 4 de octubre", "18:00–19:00",
			"$ 1.500", "$ 12.500,50", "https://app.example.com/mis-reservas",
		} {
			if !strings.Contains(body.content, want) {
				t.Errorf("%s body missing %q:\n%s", body.name, want, body.content)
			}
		}
	}
}

func TestNotificationService_SendBookingConfirmation_SenderErrorReleasesClaim(t *testing.T) {
	// El notifier corre después del commit del webhook: aunque el mail falle,
	// la reserva sigue confirmada. Acá solo verificamos que se libera el claim
	// para permitir un reintento y que el error se propaga.
	repo := &fakeNotificationRepo{detail: confirmedDetail()}
	sendErr := errors.New("resend down")
	sender := &fakeEmailSender{err: sendErr}
	svc := newNotificationSvc(repo, sender)

	err := svc.SendBookingConfirmation(context.Background(), 7)
	if !errors.Is(err, sendErr) {
		t.Fatalf("got %v, want wrapped sender error", err)
	}
	if repo.releases != 1 {
		t.Errorf("got %d releases, want 1", repo.releases)
	}
	if repo.claimed {
		t.Error("claim should be released after a failed send")
	}

	// Un reintento posterior puede enviar.
	sender.err = nil
	if err := svc.SendBookingConfirmation(context.Background(), 7); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if len(sender.sent) != 1 {
		t.Errorf("got %d emails after retry, want 1", len(sender.sent))
	}
}

func TestNotificationService_SendBookingConfirmation_NoRecipientDoesNotClaim(t *testing.T) {
	blank := "  "
	cases := map[string]*string{"nil email": nil, "blank email": &blank}
	for name, email := range cases {
		t.Run(name, func(t *testing.T) {
			d := confirmedDetail()
			d.CustomerUserEmail = email
			repo := &fakeNotificationRepo{detail: d}
			sender := &fakeEmailSender{}
			svc := newNotificationSvc(repo, sender)

			if err := svc.SendBookingConfirmation(context.Background(), 7); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(sender.sent) != 0 {
				t.Errorf("got %d emails, want 0", len(sender.sent))
			}
			if repo.claimCalls != 0 {
				t.Errorf("got %d claim calls, want 0", repo.claimCalls)
			}
		})
	}
}

func TestNotificationService_SendBookingConfirmation_ClaimErrorReleasesClaim(t *testing.T) {
	// El UPDATE pudo aplicarse aunque se perdiera la respuesta: se libera.
	claimErr := errors.New("connection reset")
	repo := &fakeNotificationRepo{detail: confirmedDetail(), claimErr: claimErr}
	sender := &fakeEmailSender{}
	svc := newNotificationSvc(repo, sender)

	err := svc.SendBookingConfirmation(context.Background(), 7)
	if !errors.Is(err, claimErr) {
		t.Fatalf("got %v, want wrapped claim error", err)
	}
	if repo.releases != 1 {
		t.Errorf("got %d releases, want 1", repo.releases)
	}
	if sender.calls != 0 {
		t.Errorf("sender called %d times, want 0", sender.calls)
	}
}

func TestNotificationService_SendBookingConfirmation_EscapesHTML(t *testing.T) {
	const hostile = "<script>alert(1)</script>"
	d := confirmedDetail()
	d.SpaceName = hostile
	d.CustomerUserName = &[]string{hostile}[0]
	repo := &fakeNotificationRepo{detail: d}
	sender := &fakeEmailSender{}

	if err := newNotificationSvc(repo, sender).SendBookingConfirmation(context.Background(), 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := sender.sent[0].html
	if strings.Contains(html, "<script>") {
		t.Errorf("html body contains unescaped script tag:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped script tag in html body:\n%s", html)
	}
}

func TestNotificationService_SendBookingConfirmation_SetsIdempotencyKey(t *testing.T) {
	repo := &fakeNotificationRepo{detail: confirmedDetail()}
	sender := &fakeEmailSender{}

	if err := newNotificationSvc(repo, sender).SendBookingConfirmation(context.Background(), 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.keys) != 1 || sender.keys[0] != "booking-confirmation-7" {
		t.Errorf("got idempotency keys %v, want [booking-confirmation-7]", sender.keys)
	}
}

func TestNotificationService_SendBookingConfirmation_BookingNotFound(t *testing.T) {
	svc := newNotificationSvc(&fakeNotificationRepo{}, &fakeEmailSender{})

	err := svc.SendBookingConfirmation(context.Background(), 7)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestNotificationService_SendBookingConfirmation_FallsBackToSlotLabel(t *testing.T) {
	d := confirmedDetail()
	d.SlotStartTime, d.SlotEndTime = nil, nil
	d.BalanceAmount = 0
	repo := &fakeNotificationRepo{detail: d}
	sender := &fakeEmailSender{}

	if err := newNotificationSvc(repo, sender).SendBookingConfirmation(context.Background(), 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sender.sent[0].text, "18 a 19") {
		t.Errorf("expected slot label in body:\n%s", sender.sent[0].text)
	}
	if strings.Contains(sender.sent[0].text, "Saldo") {
		t.Errorf("no balance line expected when balance is zero:\n%s", sender.sent[0].text)
	}
}

func TestFormatSpanishDate(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		{"saturday", time.Date(2025, time.October, 4, 0, 0, 0, 0, time.UTC), "sábado 4 de octubre"},
		{"sunday new year", time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC), "domingo 1 de enero"},
		{"wednesday", time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC), "miércoles 31 de diciembre"},
		{"leap day", time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC), "jueves 29 de febrero"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatSpanishDate(tc.in); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatPesos(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "$ 0"},
		{150, "$ 150"},
		{1500, "$ 1.500"},
		{12500, "$ 12.500"},
		{1234567, "$ 1.234.567"},
		{12500.5, "$ 12.500,50"},
		{99.05, "$ 99,05"},
		{999.999, "$ 1.000"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := formatPesos(tc.in); got != tc.want {
				t.Errorf("formatPesos(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestClockHHMM(t *testing.T) {
	cases := map[string]string{
		"18:00:00":             "18:00",
		"24:00:00":             "24:00",
		"09:30":                "09:30",
		"0000-01-01T18:00:00Z": "18:00",
		"garbage":              "",
	}
	for in, want := range cases {
		if got := clockHHMM(in); got != want {
			t.Errorf("clockHHMM(%q) = %q, want %q", in, got, want)
		}
	}
}
