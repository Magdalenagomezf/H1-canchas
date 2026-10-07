package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/database"
	"H1-canchas/pkg/mercadopago"
)

// PaymentRepo es la interfaz que el service necesita para persistir
// intentos de pago. La define el service, la implementa el repository.
type PaymentRepo interface {
	Create(ctx context.Context, tx database.Tx, p *domain.Payment) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Payment, error)
	GetLatestByBookingID(ctx context.Context, bookingID int64, kind string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, tx database.Tx, id int64, status string, rawDetail *string, externalPaymentID *string) (bool, error)
}

// BookingRepoForPayment es lo mínimo que el payment service necesita
// saber de bookings.
type BookingRepoForPayment interface {
	BeginTx(ctx context.Context) (database.Tx, error)
	GetByID(ctx context.Context, id int64) (*domain.Booking, error)
	ConfirmAfterPayment(ctx context.Context, tx database.Tx, bookingID int64, kind string) (bool, error)
	MarkPaidManually(ctx context.Context, id int64, kind, method string, recordedBy int64) (bool, error)
	MarkUnpaidManually(ctx context.Context, id int64, kind string) (bool, error)
}

// MercadoPagoClient es la porción de pkg/mercadopago que el service necesita.
type MercadoPagoClient interface {
	CreatePreference(ctx context.Context, req mercadopago.CreatePreferenceRequest) (*mercadopago.PreferenceResponse, error)
	GetPayment(ctx context.Context, paymentID int64) (*mercadopago.PaymentInfo, error)
}

// BookingConfirmationNotifier avisa que una reserva quedó confirmada tras
// el pago de la seña. La implementación debe ser no bloqueante.
type BookingConfirmationNotifier interface {
	NotifyBookingConfirmed(bookingID int64)
}

type PaymentService struct {
	repo          PaymentRepo
	bookingRepo   BookingRepoForPayment
	mp            MercadoPagoClient
	webhookSecret string
	backURLBase   string
	webhookURL    string
	notifier      BookingConfirmationNotifier // puede ser nil
}

func NewPaymentService(repo PaymentRepo, bookingRepo BookingRepoForPayment, mp MercadoPagoClient, webhookSecret, backURLBase, webhookURL string, notifier BookingConfirmationNotifier) *PaymentService {
	return &PaymentService{
		repo:          repo,
		bookingRepo:   bookingRepo,
		mp:            mp,
		webhookSecret: webhookSecret,
		backURLBase:   backURLBase,
		webhookURL:    webhookURL,
		notifier:      notifier,
	}
}

func isValidPaymentKind(kind string) bool {
	return kind == domain.PaymentKindDeposit || kind == domain.PaymentKindBalance
}

func isValidPaymentMethod(method string) bool {
	switch method {
	case domain.PaymentMethodCash, domain.PaymentMethodTransfer, domain.PaymentMethodPosnet, domain.PaymentMethodMercadoPago:
		return true
	default:
		return false
	}
}

// GenerateForBooking crea una preferencia de pago de Mercado Pago para
// un tramo (seña o saldo) de una reserva y registra el intento en payments.
func (s *PaymentService) GenerateForBooking(ctx context.Context, bookingID int64, kind string, requesterID int64, requesterRole string) (*mercadopago.PreferenceResponse, error) {
	if !isValidPaymentKind(kind) {
		return nil, ErrInvalidPaymentKind
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("PaymentService.GenerateForBooking: %w", err)
	}
	if booking == nil {
		return nil, ErrNotFound
	}

	isOwner := booking.CustomerUserID != nil && *booking.CustomerUserID == requesterID
	isStaff := requesterRole == domain.RoleReceptionist || requesterRole == domain.RoleAdmin
	if !isOwner && !isStaff {
		return nil, ErrUnauthorized
	}

	if booking.Status == domain.BookingStatusExpired {
		return nil, ErrBookingExpired
	}
	if booking.Status == domain.BookingStatusCancelled || booking.Status == domain.BookingStatusCompleted {
		return nil, ErrBookingNotPayable
	}

	var status string
	var amount float64
	if kind == domain.PaymentKindDeposit {
		status = booking.DepositStatus
		amount = booking.DepositAmount
	} else {
		status = booking.BalanceStatus
		amount = booking.BalanceAmount
	}
	if status == domain.PaymentStatusPaid {
		return nil, ErrPaymentAlreadyCompleted
	}

	// Deviation from the original plan: ExternalReference was meant to be
	// payments.id, but paymentRepo.Create has no way to attach preference_id
	// after insert, so the payments row must exist before CreatePreference is
	// called — and it can't, since its id doesn't exist yet. "bookingID:kind"
	// sidesteps the ordering problem and is still enough to correlate the
	// webhook back to a booking+kind pair.
	externalRef := fmt.Sprintf("%d:%s", bookingID, kind)
	resultURL := s.backURLBase + "/pago/resultado/" + strconv.FormatInt(bookingID, 10)

	resp, err := s.mp.CreatePreference(ctx, mercadopago.CreatePreferenceRequest{
		Title:             fmt.Sprintf("Reserva #%d — %s", bookingID, kind),
		Amount:            amount,
		ExternalReference: externalRef,
		BackURLs: mercadopago.BackURLs{
			Success: resultURL,
			Pending: resultURL,
			Failure: resultURL,
		},
		NotificationURL: s.webhookURL,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMercadoPagoUnavailable, err)
	}

	tx, err := s.bookingRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("PaymentService.GenerateForBooking: abriendo tx: %w", err)
	}
	defer tx.Rollback()

	payment := &domain.Payment{
		BookingID:    bookingID,
		Kind:         kind,
		Provider:     domain.PaymentProviderMercadoPago,
		PreferenceID: &resp.ID,
		Status:       domain.PaymentTxStatusPending,
		Amount:       amount,
	}
	if _, err := s.repo.Create(ctx, tx, payment); err != nil {
		return nil, fmt.Errorf("PaymentService.GenerateForBooking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("PaymentService.GenerateForBooking: commit: %w", err)
	}

	return resp, nil
}

// HandleWebhook procesa una notificación de Mercado Pago. Nunca confía en
// el cuerpo/query del webhook: siempre reconsulta el pago real vía
// GetPayment (con nuestro propio access token) antes de actualizar estado.
// Checkout Pro entrega estas notificaciones por el mecanismo IPN legacy, que
// según la documentación oficial de MP no soporta validación de origen por
// x-signature — por eso un mismatch queda solo logueado, no rechaza la
// request; la barrera de seguridad real es GetPayment + external_reference,
// que el atacante no puede falsificar.
func (s *PaymentService) HandleWebhook(ctx context.Context, xSignature, xRequestID, dataID string) error {
	if !mercadopago.VerifyWebhookSignature(s.webhookSecret, xSignature, xRequestID, dataID) {
		log.Printf("WARN webhook: firma no coincide para dataID=%s (esperado en notificaciones IPN legacy de Checkout Pro)", dataID)
	}

	paymentID, err := strconv.ParseInt(dataID, 10, 64)
	if err != nil {
		return fmt.Errorf("PaymentService.HandleWebhook: dataID inválido: %w", err)
	}

	info, err := s.mp.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("PaymentService.HandleWebhook: %w", err)
	}

	parts := strings.SplitN(info.ExternalReference, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("PaymentService.HandleWebhook: external_reference inválido: %q", info.ExternalReference)
	}
	bookingID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("PaymentService.HandleWebhook: external_reference inválido: %q", info.ExternalReference)
	}
	kind := parts[1]

	payment, err := s.repo.GetLatestByBookingID(ctx, bookingID, kind)
	if err != nil {
		return fmt.Errorf("PaymentService.HandleWebhook: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("PaymentService.HandleWebhook: no hay pago local para booking %d kind %q", bookingID, kind)
	}

	tx, err := s.bookingRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("PaymentService.HandleWebhook: abriendo tx: %w", err)
	}
	defer tx.Rollback()

	externalPaymentIDStr := strconv.FormatInt(info.ID, 10)
	if _, err := s.repo.UpdateStatus(ctx, tx, payment.ID, info.Status, &info.StatusDetail, &externalPaymentIDStr); err != nil {
		return fmt.Errorf("PaymentService.HandleWebhook: %w", err)
	}

	confirmed := false
	if info.Status == domain.PaymentTxStatusApproved {
		// ConfirmAfterPayment's SQL is naturally idempotent (guarded by
		// status NOT IN cancelled/expired, CASE-based status transition),
		// so no extra "already processed" guard is needed here even though
		// the same webhook can legitimately be retried by Mercado Pago.
		confirmed, err = s.bookingRepo.ConfirmAfterPayment(ctx, tx, bookingID, kind)
		if err != nil {
			return fmt.Errorf("PaymentService.HandleWebhook: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("PaymentService.HandleWebhook: commit: %w", err)
	}

	// Solo después del commit: si el mail falla, la reserva sigue confirmada.
	// Webhook duplicados no duplican el mail (claim atómico en el notifier).
	if s.notifier != nil && confirmed && kind == domain.PaymentKindDeposit {
		s.notifier.NotifyBookingConfirmed(bookingID)
	}

	return nil
}

// MarkPaidManually registra el pago de un tramo hecho fuera de Mercado Pago
// (efectivo, transferencia, posnet).
func (s *PaymentService) MarkPaidManually(ctx context.Context, bookingID int64, kind, method string, recordedByID int64) error {
	if !isValidPaymentKind(kind) {
		return ErrInvalidPaymentKind
	}
	if !isValidPaymentMethod(method) {
		return ErrInvalidPaymentMethod
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("PaymentService.MarkPaidManually: %w", err)
	}
	if booking == nil {
		return ErrNotFound
	}

	var status, amount = booking.DepositStatus, booking.DepositAmount
	if kind == domain.PaymentKindBalance {
		status, amount = booking.BalanceStatus, booking.BalanceAmount
	}
	if status == domain.PaymentStatusPaid {
		return ErrPaymentAlreadyCompleted
	}

	// bookingRepo.MarkPaidManually takes no tx (frozen repository contract
	// this phase), so it can't share a transaction with paymentRepo.Create.
	// Flipping the booking status first means that if the audit-log insert
	// below fails, the booking is still correctly marked paid — the safer
	// failure mode than leaving money collected but unrecorded as paid.
	updated, err := s.bookingRepo.MarkPaidManually(ctx, bookingID, kind, method, recordedByID)
	if err != nil {
		return fmt.Errorf("PaymentService.MarkPaidManually: %w", err)
	}
	if !updated {
		return ErrNotFound
	}

	tx, err := s.bookingRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("PaymentService.MarkPaidManually: abriendo tx para auditoría: %w", err)
	}
	defer tx.Rollback()

	payment := &domain.Payment{
		BookingID: bookingID,
		Kind:      kind,
		Provider:  domain.PaymentProviderManual,
		Status:    domain.PaymentTxStatusApproved,
		Amount:    amount,
	}
	if _, err := s.repo.Create(ctx, tx, payment); err != nil {
		return fmt.Errorf("PaymentService.MarkPaidManually: registrando auditoría: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("PaymentService.MarkPaidManually: commit: %w", err)
	}

	return nil
}

// MarkUnpaidManually revierte un tramo marcado como pagado por error.
// No genera fila de auditoría: revertir un error no es en sí un evento de pago.
func (s *PaymentService) MarkUnpaidManually(ctx context.Context, bookingID int64, kind string) error {
	if !isValidPaymentKind(kind) {
		return ErrInvalidPaymentKind
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("PaymentService.MarkUnpaidManually: %w", err)
	}
	if booking == nil {
		return ErrNotFound
	}

	updated, err := s.bookingRepo.MarkUnpaidManually(ctx, bookingID, kind)
	if err != nil {
		return fmt.Errorf("PaymentService.MarkUnpaidManually: %w", err)
	}
	if !updated {
		return ErrNotFound
	}

	return nil
}
