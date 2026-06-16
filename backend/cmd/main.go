package main

import (
	"log"
	"net/http"

	"H1-canchas/config"
	"H1-canchas/internal/controllers"
	"H1-canchas/internal/domain"
	"H1-canchas/internal/middleware"
	"H1-canchas/internal/repository"
	"H1-canchas/internal/service"
	"H1-canchas/pkg/database"

	"github.com/gin-gonic/gin"
)

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

	bookingRepo := repository.NewBookingRepository(db)
	bookingService := service.NewBookingService(bookingRepo, spaceRepo)
	bookingCtrl := controllers.NewBookingController(bookingService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
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

	// Rutas protegidas — requieren JWT válido
	protected := r.Group("/")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		protected.POST("/admin/users", middleware.RequireRole(domain.RoleAdmin), authCtrl.CreateStaff)
		protected.POST("/spaces", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), spaceController.Create)
		protected.PUT("/spaces/:id", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), spaceController.Update)
		protected.POST("/spaces/:id/slots", middleware.RequireRole(domain.RoleAdmin), spaceController.CreateSlot)
		protected.DELETE("/spaces/:id", middleware.RequireRole(domain.RoleAdmin), spaceController.Deactivate)
		// Bookings
		protected.GET("/bookings", bookingCtrl.GetMyBookings)
		protected.GET("/bookings/:id", bookingCtrl.GetByID)
		protected.POST("/bookings/manual", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), bookingCtrl.CreateManual)
		protected.POST("/bookings", bookingCtrl.Create)
		protected.PATCH("/bookings/:id/cancel", bookingCtrl.Cancel)
	}

	log.Println("Servidor corriendo en puerto " + cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
