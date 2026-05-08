package main

import (
	"log"
	"net/http"

	"H1-canchas/config"
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

	r := gin.Default()

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
