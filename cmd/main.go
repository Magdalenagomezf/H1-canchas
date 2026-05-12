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
	// te amoooooo
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	authCtrl := controllers.NewAuthController(userService)

	spaceRepo := repository.NewSpaceRepository(db)
	spaceService := service.NewSpaceService(spaceRepo)
	spaceController := controllers.NewSpaceController(spaceService)

	r := gin.Default()

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
		protected.POST("/spaces", middleware.RequireRole(domain.RoleAdmin, domain.RoleReceptionist), spaceController.Create)
		protected.DELETE("/spaces/:id", middleware.RequireRole(domain.RoleAdmin), spaceController.Deactivate)
	}

	log.Println("Servidor corriendo en puerto " + cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
