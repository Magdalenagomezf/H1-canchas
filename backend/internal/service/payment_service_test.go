package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/database"
	"H1-canchas/pkg/mercadopago"
)

// validXSignature builds an x-signature header that
// mercadopago.VerifyWebhookSignature accepts, mirroring its manifest
// format (id:<dataID>;request-id:<requestID>;ts:<ts>;) so webhook tests
// don't need a real Mercado Pago secret.
func validXSignature(secret, dataID, requestID string) string {
	const ts = "1700000000"
	manifest := "id:" + dataID + ";request-id:" + requestID + ";ts:" + ts + ";"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	v1 := hex.EncodeToString(mac.Sum(nil))
	return "ts=" + ts + ",v1=" + v1
}

// mockPaymentRepo implements PaymentRepo for unit tests.
type mockPaymentRepo struct {
	createFn               func(ctx context.Context, tx database.Tx, p *domain.Payment) (int64, error)
	getByIDFn              func(ctx context.Context, id int64) (*domain.Payment, error)
	getLatestByBookingIDFn func(ctx context.Context, bookingID int64, kind string) (*domain.Payment, error)
	updateStatusFn         func(ctx context.Context, tx database.Tx, id int64, status string, rawDetail, externalPaymentID *string) (bool, error)
	callLog                []string
}

func (m *mockPaymentRepo) Create(ctx context.Context, tx database.Tx, p *domain.Payment) (int64, error) {
	m.callLog = append(m.callLog, "paymentRepo.Create")
	if m.createFn != nil {
		return m.createFn(ctx, tx, p)
	}
	return 1, nil
}
func (m *mockPaymentRepo) GetByID(ctx context.Context, id int64) (*domain.Payment, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockPaymentRepo) GetLatestByBookingID(ctx context.Context, bookingID int64, kind string) (*domain.Payment, error) {
	if m.getLatestByBookingIDFn != nil {
		return m.getLatestByBookingIDFn(ctx, bookingID, kind)
	}
	return nil, nil
}
func (m *mockPaymentRepo) UpdateStatus(ctx context.Context, tx database.Tx, id int64, status string, rawDetail, externalPaymentID *string) (bool, error) {
	m.callLog = append(m.callLog, "paymentRepo.UpdateStatus")
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, tx, id, status, rawDetail, externalPaymentID)
	}
	return true, nil
}

// mockBookingRepoForPayment implements BookingRepoForPayment for unit tests.
type mockBookingRepoForPayment struct {
	beginTxFn             func(ctx context.Context) (database.Tx, error)
	getByIDFn             func(ctx context.Context, id int64) (*domain.Booking, error)
	confirmAfterPaymentFn func(ctx context.Context, tx database.Tx, bookingID int64, kind string) (bool, error)
	markPaidManuallyFn    func(ctx context.Context, id int64, kind, method string, recordedBy int64) (bool, error)
	markUnpaidManuallyFn  func(ctx context.Context, id int64, kind string) (bool, error)
	callLog               []string
}

func (m *mockBookingRepoForPayment) BeginTx(ctx context.Context) (database.Tx, error) {
	if m.beginTxFn != nil {
		return m.beginTxFn(ctx)
	}
	return &mockTx{}, nil
}
func (m *mockBookingRepoForPayment) GetByID(ctx context.Context, id int64) (*domain.Booking, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockBookingRepoForPayment) ConfirmAfterPayment(ctx context.Context, tx database.Tx, bookingID int64, kind string) (bool, error) {
	m.callLog = append(m.callLog, "bookingRepo.ConfirmAfterPayment")
	if m.confirmAfterPaymentFn != nil {
		return m.confirmAfterPaymentFn(ctx, tx, bookingID, kind)
	}
	return true, nil
}
func (m *mockBookingRepoForPayment) MarkPaidManually(ctx context.Context, id int64, kind, method string, recordedBy int64) (bool, error) {
	m.callLog = append(m.callLog, "bookingRepo.MarkPaidManually")
	if m.markPaidManuallyFn != nil {
		return m.markPaidManuallyFn(ctx, id, kind, method, recordedBy)
	}
	return true, nil
}
func (m *mockBookingRepoForPayment) MarkUnpaidManually(ctx context.Context, id int64, kind string) (bool, error) {
	if m.markUnpaidManuallyFn != nil {
		return m.markUnpaidManuallyFn(ctx, id, kind)
	}
	return true, nil
}

// mockMercadoPagoClient implements MercadoPagoClient for unit tests.
type mockMercadoPagoClient struct {
	createPreferenceFn func(ctx context.Context, req mercadopago.CreatePreferenceRequest) (*mercadopago.PreferenceResponse, error)
	getPaymentFn       func(ctx context.Context, paymentID int64) (*mercadopago.PaymentInfo, error)
	callLog            []string
}

func (m *mockMercadoPagoClient) CreatePreference(ctx context.Context, req mercadopago.CreatePreferenceRequest) (*mercadopago.PreferenceResponse, error) {
	m.callLog = append(m.callLog, "mp.CreatePreference")
	if m.createPreferenceFn != nil {
		return m.createPreferenceFn(ctx, req)
	}
	return &mercadopago.PreferenceResponse{ID: "pref-1", InitPoint: "https://mp/init"}, nil
}
func (m *mockMercadoPagoClient) GetPayment(ctx context.Context, paymentID int64) (*mercadopago.PaymentInfo, error) {
	if m.getPaymentFn != nil {
		return m.getPaymentFn(ctx, paymentID)
	}
	return nil, nil
}

const (
	testWebhookSecret = "test-secret"
	testBackURLBase   = "https://example.com"
	testWebhookURL    = "https://example.com/payments/webhook"
)

func newPaymentSvc(repo *mockPaymentRepo, bookingRepo *mockBookingRepoForPayment, mp *mockMercadoPagoClient) *PaymentService {
	return NewPaymentService(repo, bookingRepo, mp, testWebhookSecret, testBackURLBase, testWebhookURL, nil)
}

// mockNotifier implements BookingConfirmationNotifier for unit tests.
type mockNotifier struct {
	calls []int64
}

func (m *mockNotifier) NotifyBookingConfirmed(bookingID int64) {
	m.calls = append(m.calls, bookingID)
}

func payableBooking() *domain.Booking {
	owner := int64ptr(10)
	return &domain.Booking{
		ID:             1,
		CustomerUserID: owner,
		Status:         domain.BookingStatusPending,
		DepositAmount:  150,
		DepositStatus:  domain.PaymentStatusUnpaid,
		BalanceAmount:  850,
		BalanceStatus:  domain.PaymentStatusUnpaid,
	}
}

// --- GenerateForBooking ---

func TestPaymentService_GenerateForBooking_InvalidKind(t *testing.T) {
	svc := newPaymentSvc(&mockPaymentRepo{}, &mockBookingRepoForPayment{}, &mockMercadoPagoClient{})

	_, err := svc.GenerateForBooking(context.Background(), 1, "invalid", 10, domain.RoleCustomer)
	if !errors.Is(err, ErrInvalidPaymentKind) {
		t.Errorf("got %v, want ErrInvalidPaymentKind", err)
	}
}

func TestPaymentService_GenerateForBooking_BookingNotFound(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return nil, nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	_, err := svc.GenerateForBooking(context.Background(), 1, domain.PaymentKindDeposit, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestPaymentService_GenerateForBooking_Unauthorized(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	// requesterID=99 is not the owner (10) and not staff
	_, err := svc.GenerateForBooking(context.Background(), 1, domain.PaymentKindDeposit, 99, domain.RoleCustomer)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("got %v, want ErrUnauthorized", err)
	}
}

func TestPaymentService_GenerateForBooking_BookingExpired(t *testing.T) {
	booking := payableBooking()
	booking.Status = domain.BookingStatusExpired
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	_, err := svc.GenerateForBooking(context.Background(), 1, domain.PaymentKindDeposit, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrBookingExpired) {
		t.Errorf("got %v, want ErrBookingExpired", err)
	}
}

func TestPaymentService_GenerateForBooking_BookingNotPayable(t *testing.T) {
	cases := []string{domain.BookingStatusCancelled, domain.BookingStatusCompleted}
	for _, status := range cases {
		t.Run(status, func(t *testing.T) {
			booking := payableBooking()
			booking.Status = status
			bookingRepo := &mockBookingRepoForPayment{
				getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
			}
			svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

			_, err := svc.GenerateForBooking(context.Background(), 1, domain.PaymentKindDeposit, 10, domain.RoleCustomer)
			if !errors.Is(err, ErrBookingNotPayable) {
				t.Errorf("status %s: got %v, want ErrBookingNotPayable", status, err)
			}
		})
	}
}

func TestPaymentService_GenerateForBooking_AlreadyPaid(t *testing.T) {
	cases := []struct {
		kind string
	}{
		{domain.PaymentKindDeposit},
		{domain.PaymentKindBalance},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			booking := payableBooking()
			if tc.kind == domain.PaymentKindDeposit {
				booking.DepositStatus = domain.PaymentStatusPaid
			} else {
				booking.BalanceStatus = domain.PaymentStatusPaid
			}
			bookingRepo := &mockBookingRepoForPayment{
				getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
			}
			svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

			_, err := svc.GenerateForBooking(context.Background(), 1, tc.kind, 10, domain.RoleCustomer)
			if !errors.Is(err, ErrPaymentAlreadyCompleted) {
				t.Errorf("kind %s: got %v, want ErrPaymentAlreadyCompleted", tc.kind, err)
			}
		})
	}
}

func TestPaymentService_GenerateForBooking_MercadoPagoError(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
	}
	mpErr := errors.New("mp down")
	mp := &mockMercadoPagoClient{
		createPreferenceFn: func(_ context.Context, _ mercadopago.CreatePreferenceRequest) (*mercadopago.PreferenceResponse, error) {
			return nil, mpErr
		},
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, mp)

	_, err := svc.GenerateForBooking(context.Background(), 1, domain.PaymentKindDeposit, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrMercadoPagoUnavailable) {
		t.Errorf("got %v, want ErrMercadoPagoUnavailable", err)
	}
}

func TestPaymentService_GenerateForBooking_HappyPath_OwnerOrStaff(t *testing.T) {
	cases := []struct {
		name          string
		requesterID   int64
		requesterRole string
	}{
		{"owner", 10, domain.RoleCustomer},
		{"staff", 999, domain.RoleReceptionist},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotReq mercadopago.CreatePreferenceRequest
			mp := &mockMercadoPagoClient{
				createPreferenceFn: func(_ context.Context, req mercadopago.CreatePreferenceRequest) (*mercadopago.PreferenceResponse, error) {
					gotReq = req
					return &mercadopago.PreferenceResponse{ID: "pref-123", InitPoint: "https://mp/init"}, nil
				},
			}
			var createdPayment *domain.Payment
			paymentRepo := &mockPaymentRepo{
				createFn: func(_ context.Context, _ database.Tx, p *domain.Payment) (int64, error) {
					createdPayment = p
					return 55, nil
				},
			}
			bookingRepo := &mockBookingRepoForPayment{
				getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
			}
			svc := newPaymentSvc(paymentRepo, bookingRepo, mp)

			resp, err := svc.GenerateForBooking(context.Background(), 1, domain.PaymentKindDeposit, tc.requesterID, tc.requesterRole)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.ID != "pref-123" {
				t.Errorf("got resp ID %q, want pref-123", resp.ID)
			}
			if gotReq.ExternalReference != "1:deposit" {
				t.Errorf("got external reference %q, want %q", gotReq.ExternalReference, "1:deposit")
			}
			if createdPayment == nil {
				t.Fatal("expected paymentRepo.Create to be called")
			}
			if createdPayment.PreferenceID == nil || *createdPayment.PreferenceID != "pref-123" {
				t.Errorf("got preference id %v, want pref-123", createdPayment.PreferenceID)
			}
			if createdPayment.Status != domain.PaymentTxStatusPending {
				t.Errorf("got status %q, want pending", createdPayment.Status)
			}
			if createdPayment.Provider != domain.PaymentProviderMercadoPago {
				t.Errorf("got provider %q, want mercadopago", createdPayment.Provider)
			}

			if len(mp.callLog) != 1 || mp.callLog[0] != "mp.CreatePreference" {
				t.Fatalf("expected mp.CreatePreference to be called once, got %v", mp.callLog)
			}
			if len(paymentRepo.callLog) != 1 || paymentRepo.callLog[0] != "paymentRepo.Create" {
				t.Fatalf("expected paymentRepo.Create to be called once, got %v", paymentRepo.callLog)
			}
		})
	}
}

// --- HandleWebhook ---

// Checkout Pro entrega notificaciones vía IPN legacy, que la doc oficial de
// MP indica que no soporta validación de origen por x-signature. Un mismatch
// de firma no bloquea la request — sigue procesando y confía en la
// reconsulta a GetPayment (con nuestro propio access token) como barrera de
// seguridad real, ya que el atacante no puede falsificar esa respuesta.
func TestPaymentService_HandleWebhook_SignatureMismatch_StillProcessesViaGetPayment(t *testing.T) {
	mp := &mockMercadoPagoClient{
		getPaymentFn: func(_ context.Context, _ int64) (*mercadopago.PaymentInfo, error) {
			return &mercadopago.PaymentInfo{ID: 123, Status: domain.PaymentTxStatusApproved, ExternalReference: "1:deposit"}, nil
		},
	}
	paymentRepo := &mockPaymentRepo{
		getLatestByBookingIDFn: func(_ context.Context, _ int64, _ string) (*domain.Payment, error) {
			return &domain.Payment{ID: 5, BookingID: 1, Kind: domain.PaymentKindDeposit}, nil
		},
	}
	bookingRepo := &mockBookingRepoForPayment{}
	svc := newPaymentSvc(paymentRepo, bookingRepo, mp)

	err := svc.HandleWebhook(context.Background(), "ts=1,v1=bogus", "req-1", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, c := range bookingRepo.callLog {
		if c == "bookingRepo.ConfirmAfterPayment" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ConfirmAfterPayment to be called despite signature mismatch, callLog=%v", bookingRepo.callLog)
	}
}

func validSignatureFor(dataID, requestID string) string {
	// mirrors mercadopago.VerifyWebhookSignature's manifest format so tests
	// don't need to hardcode a HMAC value.
	return validXSignature(testWebhookSecret, dataID, requestID)
}

func TestPaymentService_HandleWebhook_MalformedDataID(t *testing.T) {
	xSig := validSignatureFor("not-a-number", "req-1")
	svc := newPaymentSvc(&mockPaymentRepo{}, &mockBookingRepoForPayment{}, &mockMercadoPagoClient{})

	err := svc.HandleWebhook(context.Background(), xSig, "req-1", "not-a-number")
	if err == nil {
		t.Fatal("expected error for malformed dataID, got nil")
	}
}

func TestPaymentService_HandleWebhook_MalformedExternalReference(t *testing.T) {
	xSig := validSignatureFor("123", "req-1")
	mp := &mockMercadoPagoClient{
		getPaymentFn: func(_ context.Context, _ int64) (*mercadopago.PaymentInfo, error) {
			return &mercadopago.PaymentInfo{ID: 123, Status: domain.PaymentTxStatusApproved, ExternalReference: "garbage"}, nil
		},
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, &mockBookingRepoForPayment{}, mp)

	err := svc.HandleWebhook(context.Background(), xSig, "req-1", "123")
	if err == nil {
		t.Fatal("expected error for malformed external reference, got nil")
	}
}

func TestPaymentService_HandleWebhook_NoMatchingLocalPayment(t *testing.T) {
	xSig := validSignatureFor("123", "req-1")
	mp := &mockMercadoPagoClient{
		getPaymentFn: func(_ context.Context, _ int64) (*mercadopago.PaymentInfo, error) {
			return &mercadopago.PaymentInfo{ID: 123, Status: domain.PaymentTxStatusApproved, ExternalReference: "1:deposit"}, nil
		},
	}
	paymentRepo := &mockPaymentRepo{
		getLatestByBookingIDFn: func(_ context.Context, _ int64, _ string) (*domain.Payment, error) { return nil, nil },
	}
	svc := newPaymentSvc(paymentRepo, &mockBookingRepoForPayment{}, mp)

	err := svc.HandleWebhook(context.Background(), xSig, "req-1", "123")
	if err == nil {
		t.Fatal("expected error for missing local payment, got nil")
	}
}

func TestPaymentService_HandleWebhook_Approved_ConfirmsBooking(t *testing.T) {
	xSig := validSignatureFor("123", "req-1")
	mp := &mockMercadoPagoClient{
		getPaymentFn: func(_ context.Context, _ int64) (*mercadopago.PaymentInfo, error) {
			return &mercadopago.PaymentInfo{ID: 123, Status: domain.PaymentTxStatusApproved, ExternalReference: "1:deposit"}, nil
		},
	}
	paymentRepo := &mockPaymentRepo{
		getLatestByBookingIDFn: func(_ context.Context, _ int64, _ string) (*domain.Payment, error) {
			return &domain.Payment{ID: 5, BookingID: 1, Kind: domain.PaymentKindDeposit}, nil
		},
	}
	bookingRepo := &mockBookingRepoForPayment{}
	svc := newPaymentSvc(paymentRepo, bookingRepo, mp)

	err := svc.HandleWebhook(context.Background(), xSig, "req-1", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, c := range bookingRepo.callLog {
		if c == "bookingRepo.ConfirmAfterPayment" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ConfirmAfterPayment to be called, callLog=%v", bookingRepo.callLog)
	}
}

func TestPaymentService_HandleWebhook_Rejected_DoesNotConfirmBooking(t *testing.T) {
	xSig := validSignatureFor("123", "req-1")
	mp := &mockMercadoPagoClient{
		getPaymentFn: func(_ context.Context, _ int64) (*mercadopago.PaymentInfo, error) {
			return &mercadopago.PaymentInfo{ID: 123, Status: domain.PaymentTxStatusRejected, ExternalReference: "1:deposit"}, nil
		},
	}
	paymentRepo := &mockPaymentRepo{
		getLatestByBookingIDFn: func(_ context.Context, _ int64, _ string) (*domain.Payment, error) {
			return &domain.Payment{ID: 5, BookingID: 1, Kind: domain.PaymentKindDeposit}, nil
		},
	}
	bookingRepo := &mockBookingRepoForPayment{}
	svc := newPaymentSvc(paymentRepo, bookingRepo, mp)

	err := svc.HandleWebhook(context.Background(), xSig, "req-1", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range bookingRepo.callLog {
		if c == "bookingRepo.ConfirmAfterPayment" {
			t.Errorf("expected ConfirmAfterPayment NOT to be called, callLog=%v", bookingRepo.callLog)
		}
	}

	if len(paymentRepo.callLog) != 1 || paymentRepo.callLog[0] != "paymentRepo.UpdateStatus" {
		t.Errorf("expected paymentRepo.UpdateStatus to be called, callLog=%v", paymentRepo.callLog)
	}
}

// --- MarkPaidManually ---

func TestPaymentService_MarkPaidManually_InvalidKind(t *testing.T) {
	svc := newPaymentSvc(&mockPaymentRepo{}, &mockBookingRepoForPayment{}, &mockMercadoPagoClient{})

	err := svc.MarkPaidManually(context.Background(), 1, "invalid", domain.PaymentMethodCash, 1)
	if !errors.Is(err, ErrInvalidPaymentKind) {
		t.Errorf("got %v, want ErrInvalidPaymentKind", err)
	}
}

func TestPaymentService_MarkPaidManually_InvalidMethod(t *testing.T) {
	svc := newPaymentSvc(&mockPaymentRepo{}, &mockBookingRepoForPayment{}, &mockMercadoPagoClient{})

	err := svc.MarkPaidManually(context.Background(), 1, domain.PaymentKindDeposit, "invalid", 1)
	if !errors.Is(err, ErrInvalidPaymentMethod) {
		t.Errorf("got %v, want ErrInvalidPaymentMethod", err)
	}
}

func TestPaymentService_MarkPaidManually_BookingNotFound(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return nil, nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkPaidManually(context.Background(), 1, domain.PaymentKindDeposit, domain.PaymentMethodCash, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestPaymentService_MarkPaidManually_AlreadyPaid(t *testing.T) {
	booking := payableBooking()
	booking.DepositStatus = domain.PaymentStatusPaid
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkPaidManually(context.Background(), 1, domain.PaymentKindDeposit, domain.PaymentMethodCash, 1)
	if !errors.Is(err, ErrPaymentAlreadyCompleted) {
		t.Errorf("got %v, want ErrPaymentAlreadyCompleted", err)
	}
}

func TestPaymentService_MarkPaidManually_BookingRepoReturnsFalse(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn:          func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
		markPaidManuallyFn: func(_ context.Context, _ int64, _, _ string, _ int64) (bool, error) { return false, nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkPaidManually(context.Background(), 1, domain.PaymentKindDeposit, domain.PaymentMethodCash, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestPaymentService_MarkPaidManually_HappyPath_OrderOfWrites(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
	}
	var createdPayment *domain.Payment
	paymentRepo := &mockPaymentRepo{
		createFn: func(_ context.Context, _ database.Tx, p *domain.Payment) (int64, error) {
			createdPayment = p
			return 1, nil
		},
	}
	svc := newPaymentSvc(paymentRepo, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkPaidManually(context.Background(), 1, domain.PaymentKindDeposit, domain.PaymentMethodCash, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bookingRepo.callLog) != 1 || bookingRepo.callLog[0] != "bookingRepo.MarkPaidManually" {
		t.Fatalf("expected bookingRepo.MarkPaidManually to be called once, got %v", bookingRepo.callLog)
	}
	if len(paymentRepo.callLog) != 1 || paymentRepo.callLog[0] != "paymentRepo.Create" {
		t.Fatalf("expected paymentRepo.Create to be called once, got %v", paymentRepo.callLog)
	}
	if createdPayment == nil {
		t.Fatal("expected a payment audit row to be created")
	}
	if createdPayment.Provider != domain.PaymentProviderManual {
		t.Errorf("got provider %q, want manual", createdPayment.Provider)
	}
	if createdPayment.Status != domain.PaymentTxStatusApproved {
		t.Errorf("got status %q, want approved", createdPayment.Status)
	}
	if createdPayment.Amount != 150 {
		t.Errorf("got amount %v, want 150 (deposit amount)", createdPayment.Amount)
	}
}

func TestPaymentService_MarkPaidManually_AuditInsertFailureDoesNotRollBackBookingFlip(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
	}
	auditErr := errors.New("audit insert failed")
	paymentRepo := &mockPaymentRepo{
		createFn: func(_ context.Context, _ database.Tx, _ *domain.Payment) (int64, error) { return 0, auditErr },
	}
	svc := newPaymentSvc(paymentRepo, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkPaidManually(context.Background(), 1, domain.PaymentKindDeposit, domain.PaymentMethodCash, 42)
	if err == nil {
		t.Fatal("expected error from failed audit insert, got nil")
	}
	if len(bookingRepo.callLog) != 1 || bookingRepo.callLog[0] != "bookingRepo.MarkPaidManually" {
		t.Errorf("expected bookingRepo.MarkPaidManually to have already run, got %v", bookingRepo.callLog)
	}
}

// --- MarkUnpaidManually ---

func TestPaymentService_MarkUnpaidManually_InvalidKind(t *testing.T) {
	svc := newPaymentSvc(&mockPaymentRepo{}, &mockBookingRepoForPayment{}, &mockMercadoPagoClient{})

	err := svc.MarkUnpaidManually(context.Background(), 1, "invalid")
	if !errors.Is(err, ErrInvalidPaymentKind) {
		t.Errorf("got %v, want ErrInvalidPaymentKind", err)
	}
}

func TestPaymentService_MarkUnpaidManually_BookingNotFound(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return nil, nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkUnpaidManually(context.Background(), 1, domain.PaymentKindDeposit)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestPaymentService_MarkUnpaidManually_RepoReturnsFalse(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn:            func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
		markUnpaidManuallyFn: func(_ context.Context, _ int64, _ string) (bool, error) { return false, nil },
	}
	svc := newPaymentSvc(&mockPaymentRepo{}, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkUnpaidManually(context.Background(), 1, domain.PaymentKindDeposit)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestPaymentService_MarkUnpaidManually_HappyPath(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return payableBooking(), nil },
	}
	paymentRepo := &mockPaymentRepo{}
	svc := newPaymentSvc(paymentRepo, bookingRepo, &mockMercadoPagoClient{})

	err := svc.MarkUnpaidManually(context.Background(), 1, domain.PaymentKindDeposit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paymentRepo.callLog) != 0 {
		t.Errorf("expected no audit row on unmark, got calls %v", paymentRepo.callLog)
	}
}

// --- HandleWebhook: notificación de confirmación ---

// webhookNotifierFixture arma un service con notifier para un pago de
// MP con el estado y external_reference indicados.
func webhookNotifierFixture(status, externalRef string, bookingRepo *mockBookingRepoForPayment) (*PaymentService, *mockNotifier) {
	mp := &mockMercadoPagoClient{
		getPaymentFn: func(_ context.Context, _ int64) (*mercadopago.PaymentInfo, error) {
			return &mercadopago.PaymentInfo{ID: 123, Status: status, ExternalReference: externalRef}, nil
		},
	}
	paymentRepo := &mockPaymentRepo{
		getLatestByBookingIDFn: func(_ context.Context, _ int64, _ string) (*domain.Payment, error) {
			return &domain.Payment{ID: 5, BookingID: 1}, nil
		},
	}
	notifier := &mockNotifier{}
	svc := NewPaymentService(paymentRepo, bookingRepo, mp, testWebhookSecret, testBackURLBase, testWebhookURL, notifier)
	return svc, notifier
}

func TestPaymentService_HandleWebhook_Notifier(t *testing.T) {
	commitErr := errors.New("commit failed")
	cases := []struct {
		name        string
		status      string
		externalRef string
		confirmed   bool
		commitErr   error
		wantCalls   int
		wantErr     bool
	}{
		{"approved deposit notifies once", domain.PaymentTxStatusApproved, "1:deposit", true, nil, 1, false},
		{"approved balance does not notify", domain.PaymentTxStatusApproved, "1:balance", true, nil, 0, false},
		{"rejected deposit does not notify", domain.PaymentTxStatusRejected, "1:deposit", true, nil, 0, false},
		{"approved but booking not confirmed does not notify", domain.PaymentTxStatusApproved, "1:deposit", false, nil, 0, false},
		{"commit failure does not notify", domain.PaymentTxStatusApproved, "1:deposit", true, commitErr, 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bookingRepo := &mockBookingRepoForPayment{
				beginTxFn: func(_ context.Context) (database.Tx, error) {
					return &mockTx{commitErr: tc.commitErr}, nil
				},
				confirmAfterPaymentFn: func(_ context.Context, _ database.Tx, _ int64, _ string) (bool, error) {
					return tc.confirmed, nil
				},
			}
			svc, notifier := webhookNotifierFixture(tc.status, tc.externalRef, bookingRepo)

			err := svc.HandleWebhook(context.Background(), "ts=1,v1=bogus", "req-1", "123")
			if (err != nil) != tc.wantErr {
				t.Fatalf("got err %v, wantErr %v", err, tc.wantErr)
			}
			if len(notifier.calls) != tc.wantCalls {
				t.Fatalf("got %d notifier calls, want %d", len(notifier.calls), tc.wantCalls)
			}
			if tc.wantCalls == 1 && notifier.calls[0] != 1 {
				t.Errorf("notified booking %d, want 1", notifier.calls[0])
			}
		})
	}
}

func TestPaymentService_HandleWebhook_NilNotifierIsSafe(t *testing.T) {
	bookingRepo := &mockBookingRepoForPayment{}
	svc, _ := webhookNotifierFixture(domain.PaymentTxStatusApproved, "1:deposit", bookingRepo)
	svc.notifier = nil

	if err := svc.HandleWebhook(context.Background(), "ts=1,v1=bogus", "req-1", "123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
