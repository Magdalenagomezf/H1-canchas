package service

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	htmltemplate "html/template"
	"log"
	"math"
	"strconv"
	"strings"
	texttemplate "text/template"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/email"
)

//go:embed templates/booking_confirmation.html
var bookingConfirmationHTML string

//go:embed templates/booking_confirmation.txt
var bookingConfirmationText string

var (
	bookingConfirmationHTMLTmpl = htmltemplate.Must(htmltemplate.New("booking_confirmation.html").Parse(bookingConfirmationHTML))
	bookingConfirmationTextTmpl = texttemplate.Must(texttemplate.New("booking_confirmation.txt").Parse(bookingConfirmationText))
)

// EmailSender envía un mail transaccional. Lo implementa pkg/email.
type EmailSender interface {
	Send(ctx context.Context, to, subject, html, text string) error
}

// NotificationRepo es lo que el notification service necesita de bookings.
type NotificationRepo interface {
	GetByIDWithDetails(ctx context.Context, id int64) (*domain.BookingDetail, error)
	ClaimConfirmationEmail(ctx context.Context, bookingID int64) (bool, error)
	ReleaseConfirmationEmail(ctx context.Context, bookingID int64) error
}

type NotificationService struct {
	repo        NotificationRepo
	sender      EmailSender
	frontendURL string
}

func NewNotificationService(repo NotificationRepo, sender EmailSender, frontendURL string) *NotificationService {
	return &NotificationService{
		repo:        repo,
		sender:      sender,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

type bookingConfirmationData struct {
	CustomerName  string
	SpaceName     string
	DateText      string
	TimeText      string
	DepositPaid   string
	BalanceDue    string // vacío si no hay saldo pendiente
	MyBookingsURL string
}

// NotifyBookingConfirmed dispara el mail de confirmación sin bloquear al
// llamador. Corre en su propia goroutine con un contexto propio: nunca
// debe usar el contexto del request, que se cancela al responder el webhook.
func (s *NotificationService) NotifyBookingConfirmed(bookingID int64) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.SendBookingConfirmation(ctx, bookingID); err != nil {
			log.Printf("WARN notification: mail de confirmación de la reserva %d: %v", bookingID, err)
		}
	}()
}

// SendBookingConfirmation envía (una sola vez) el mail de confirmación de
// una reserva. Es síncrono para poder testearlo.
//
// Orden: sin destinatario no se reclama nada (así, si el usuario agrega un
// email después, el envío sigue disponible); el claim atómico en DB evita
// duplicados ante reintentos del webhook; si el envío falla se libera el
// claim para permitir un nuevo intento.
func (s *NotificationService) SendBookingConfirmation(ctx context.Context, bookingID int64) error {
	detail, err := s.repo.GetByIDWithDetails(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("NotificationService.SendBookingConfirmation: %w", err)
	}
	if detail == nil {
		return ErrNotFound
	}

	if detail.CustomerUserEmail == nil || strings.TrimSpace(*detail.CustomerUserEmail) == "" {
		// Reserva manual o usuario sin email: no hay a quién avisar.
		return nil
	}
	to := strings.TrimSpace(*detail.CustomerUserEmail)

	// Known tradeoff: si el proceso muere (deploy/crash) entre el claim y el
	// envío, ese mail se pierde. Se acepta para este volumen; evitarlo
	// requeriría un estado intermedio ("sending") o una cola.
	claimed, err := s.repo.ClaimConfirmationEmail(ctx, bookingID)
	if err != nil {
		// El UPDATE pudo haberse aplicado y haberse perdido solo la
		// respuesta (timeout/conexión caída): liberar es inocuo si la
		// columna sigue en NULL y evita dejar la reserva marcada como enviada.
		s.release(bookingID)
		return fmt.Errorf("NotificationService.SendBookingConfirmation: %w", err)
	}
	if !claimed {
		return nil
	}

	// La idempotency key evita un duplicado si el cliente HTTP da timeout
	// después de que Resend aceptó el mail y el webhook se reintenta.
	ctx = email.WithIdempotencyKey(ctx, fmt.Sprintf("booking-confirmation-%d", bookingID))

	subject, htmlBody, textBody, err := s.render(detail)
	if err != nil {
		s.release(bookingID)
		return fmt.Errorf("NotificationService.SendBookingConfirmation: %w", err)
	}

	if err := s.sender.Send(ctx, to, subject, htmlBody, textBody); err != nil {
		s.release(bookingID)
		return fmt.Errorf("NotificationService.SendBookingConfirmation: enviando mail: %w", err)
	}
	return nil
}

// release libera el claim con un contexto propio: el ctx original puede
// estar vencido justamente porque el envío falló por timeout.
func (s *NotificationService) release(bookingID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.repo.ReleaseConfirmationEmail(ctx, bookingID); err != nil {
		log.Printf("WARN notification: no se pudo liberar el claim de la reserva %d: %v", bookingID, err)
	}
}

func (s *NotificationService) render(d *domain.BookingDetail) (subject, htmlBody, textBody string, err error) {
	dateText := formatSpanishDate(d.BookingDate)

	data := bookingConfirmationData{
		SpaceName:     d.SpaceName,
		DateText:      dateText,
		TimeText:      slotTimeText(d),
		DepositPaid:   formatPesos(d.DepositAmount),
		MyBookingsURL: s.frontendURL + "/mis-reservas",
	}
	if d.CustomerUserName != nil {
		data.CustomerName = strings.TrimSpace(*d.CustomerUserName)
	}
	if d.BalanceAmount > 0 && d.BalanceStatus != domain.PaymentStatusPaid {
		data.BalanceDue = formatPesos(d.BalanceAmount)
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := bookingConfirmationHTMLTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", "", fmt.Errorf("render html: %w", err)
	}
	if err := bookingConfirmationTextTmpl.Execute(&textBuf, data); err != nil {
		return "", "", "", fmt.Errorf("render text: %w", err)
	}

	subject = fmt.Sprintf("Reserva confirmada — %s, %s", d.SpaceName, dateText)
	return subject, htmlBuf.String(), textBuf.String(), nil
}

var spanishWeekdays = [...]string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"}

var spanishMonths = [...]string{
	"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

// formatSpanishDate devuelve algo como "sábado 4 de octubre".
//
// booking_date es un DATE de Postgres que se escanea como medianoche UTC.
// NO hay que convertirlo a America/Argentina/Buenos_Aires: Argentina es
// UTC-3, así que 00:00 UTC pasaría a ser las 21:00 del día anterior y el
// mail diría un día de menos. Por eso se usan Year/Month/Day/Weekday del
// propio valor, sin tocar la zona horaria, y los nombres en español salen
// de tablas escritas a mano (no dependemos de locale del sistema).
func formatSpanishDate(d time.Time) string {
	return fmt.Sprintf("%s %d de %s", spanishWeekdays[d.Weekday()], d.Day(), spanishMonths[d.Month()-1])
}

// formatPesos formatea un monto como "$ 12.500" (sin decimales si es entero)
// o "$ 12.500,50" cuando hay centavos.
func formatPesos(amount float64) string {
	cents := int64(math.Round(math.Abs(amount) * 100))
	whole, frac := cents/100, cents%100

	digits := strconv.FormatInt(whole, 10)
	var b strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}

	sign := ""
	if amount < 0 && cents > 0 {
		sign = "-"
	}
	if frac == 0 {
		return fmt.Sprintf("%s$ %s", sign, b.String())
	}
	return fmt.Sprintf("%s$ %s,%02d", sign, b.String(), frac)
}

// slotTimeText usa "HH:MM–HH:MM" si el slot tiene horarios; si no, su label.
func slotTimeText(d *domain.BookingDetail) string {
	if d.SlotStartTime != nil && d.SlotEndTime != nil {
		start, end := clockHHMM(*d.SlotStartTime), clockHHMM(*d.SlotEndTime)
		if start != "" && end != "" {
			return start + "–" + end
		}
	}
	return d.SlotLabel
}

// clockHHMM extrae "HH:MM" de un TIME de Postgres. Acepta "18:00:00",
// "24:00:00" (válido en Postgres) o un timestamp RFC3339 según cómo el
// driver haya convertido el valor.
func clockHHMM(v string) string {
	v = strings.TrimSpace(v)
	if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return t.Format("15:04")
	}
	if len(v) >= 5 && v[2] == ':' {
		return v[:5]
	}
	return ""
}
