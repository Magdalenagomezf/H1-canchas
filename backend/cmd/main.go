package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"H1-canchas/config"
	"H1-canchas/internal/controllers"
	"H1-canchas/internal/domain"
	"H1-canchas/internal/middleware"
	"H1-canchas/internal/repository"
	"H1-canchas/internal/service"
	"H1-canchas/pkg/database"
	"H1-canchas/pkg/email"
	"H1-canchas/pkg/mercadopago"

	"github.com/gin-gonic/gin"
)

// bookingExpirySweeper es el subconjunto de BookingRepository que necesita
// el sweep de background para liberar holds de reservas 'pending' vencidas.
type bookingExpirySweeper interface {
	ExpireAllStalePending(ctx context.Context) (int64, error)
}

// runBookingExpirySweep corre en background y transiciona a 'expired'
// cualquier reserva 'pending' cuyo hold (expires_at) ya venció, sin
// depender de que alguien intente reservar ese mismo slot de nuevo.
func runBookingExpirySweep(repo bookingExpirySweeper, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		count, err := repo.ExpireAllStalePending(ctx)
		cancel()
		if err != nil {
			log.Printf("booking expiry sweep: %v", err)
			continue
		}
		if count > 0 {
			log.Printf("booking expiry sweep: %d reserva(s) expiradas", count)
		}
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	authCtrl := controllers.NewAuthController(userService)

	spaceRepo := repository.NewSpaceRepository(db)
	spaceService := service.NewSpaceService(spaceRepo)
	spaceController := controllers.NewSpaceController(spaceService)

	mpClient, err := mercadopago.NewClient(cfg.MPAccessToken)
	if err != nil {
		log.Fatal(err)
	}

	bookingRepo := repository.NewBookingRepository(db)
	holdTTL := time.Duration(cfg.BookingHoldTTLMinutes) * time.Minute
	bookingService := service.NewBookingService(bookingRepo, spaceRepo, holdTTL, cfg.DepositPercentage)
	bookingCtrl := controllers.NewBookingController(bookingService)

	sweepInterval := time.Duration(cfg.BookingExpirySweepIntervalMinutes) * time.Minute
	go runBookingExpirySweep(bookingRepo, sweepInterval)

	bookingBatchRepo := repository.NewBookingBatchRepository(db)
	bookingBatchService := service.NewBookingBatchService(bookingBatchRepo, bookingRepo, spaceRepo, bookingService)
	bookingBatchCtrl := controllers.NewBookingBatchController(bookingBatchService)

	paymentRepo := repository.NewPaymentRepository(db)

	var emailSender service.EmailSender
	if cfg.ResendAPIKey == "" {
		log.Println("RESEND_API_KEY no configurada: los mails solo se loguean (destinatario y asunto), no se envían")
		emailSender = email.LogSender{}
	} else {
		emailSender = email.NewResendSender(cfg.ResendAPIKey, cfg.EmailFrom)
	}
	notificationService := service.NewNotificationService(bookingRepo, emailSender, cfg.FrontendURL)

	paymentService := service.NewPaymentService(paymentRepo, bookingRepo, mpClient, cfg.MPWebhookSecret, cfg.AllowedOrigin, cfg.MPWebhookURL, notificationService)
	paymentCtrl := controllers.NewPaymentController(paymentService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.AllowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	auth := r.Group("/auth")
	{
		auth.POST("/register", authCtrl.Register)
		auth.POST("/login", authCtrl.Login)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"db":     "connected",
		})
	})

	// Rutas públicas
	r.GET("/spaces", spaceController.GetAll)
	r.GET("/spaces/:id", spaceController.GetByID)
	r.GET("/spaces/:id/slots", spaceController.GetSlots)
	// Lo llama Mercado Pago directamente, no puede requerir JWT del front.
	r.POST("/payments/webhook", paymentCtrl.Webhook)

	// Rutas protegidas — requieren JWT válido
	protected := r.Group("/")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		protected.GET("/admin/users", middleware.RequireRole(domain.RoleAdmin), authCtrl.ListUsers)
		protected.POST("/admin/users", middleware.RequireRole(domain.RoleAdmin), authCtrl.CreateStaff)
		protected.PATCH("/admin/users/:id/role", middleware.RequireRole(domain.RoleAdmin), authCtrl.UpdateRole)
		protected.DELETE("/admin/users/:id", middleware.RequireRole(domain.RoleAdmin), authCtrl.DeleteUser)
		protected.POST("/spaces", middleware.RequireRole(domain.RoleAdmin), spaceController.Create)
		protected.PUT("/spaces/:id", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), spaceController.Update)
		protected.POST("/spaces/:id/slots", middleware.RequireRole(domain.RoleAdmin), spaceController.CreateSlot)
		protected.DELETE("/spaces/:id", middleware.RequireRole(domain.RoleAdmin), spaceController.Deactivate)
		protected.DELETE("/admin/spaces/:id", middleware.RequireRole(domain.RoleAdmin), spaceController.HardDelete)
		// Bookings
		protected.GET("/bookings", bookingCtrl.GetMyBookings)
		protected.GET("/bookings/:id", bookingCtrl.GetByID)
		protected.GET("/admin/bookings", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), bookingCtrl.GetAll)
		protected.POST("/bookings/manual", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), bookingCtrl.CreateManual)
		protected.POST("/bookings", bookingCtrl.Create)
		protected.PATCH("/bookings/:id/cancel", bookingCtrl.Cancel)
		// Turnos fijos (profesores) y bloqueos por mantenimiento
		protected.POST("/bookings/recurring", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), bookingBatchCtrl.CreateRecurring)
		protected.POST("/bookings/block", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), bookingBatchCtrl.CreateBlock)
		protected.GET("/bookings/batches", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), bookingBatchCtrl.GetAll)
		protected.PATCH("/bookings/batches/:id/cancel", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), bookingBatchCtrl.Cancel)
		// Payments
		protected.POST("/bookings/:id/payments", paymentCtrl.GeneratePreference)
		protected.PATCH("/bookings/:id/payment", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), paymentCtrl.MarkPayment)
	}

	log.Println("Servidor corriendo en puerto " + cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
