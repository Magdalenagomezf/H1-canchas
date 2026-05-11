package main

import (
	"log"
	"net/http"

	"H1-canchas/config"
	"H1-canchas/internal/controllers"
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

	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/register", authCtrl.Register)
		auth.POST("/login", authCtrl.Login)
	}

	// Rutas protegidas — requieren JWT válido
	protected := r.Group("/")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		// acá van los endpoints que agregues después
		// protected.GET("/spaces", spaceCtrl.GetAll)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"db":     "connected",
		})
	})

	log.Println("Servidor corriendo en puerto " + cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
