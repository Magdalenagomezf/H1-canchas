package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const healthDBTimeout = 2 * time.Second

// dbPinger is satisfied by *sqlx.DB.
type dbPinger interface {
	PingContext(ctx context.Context) error
}

type HealthController struct {
	db     dbPinger
	commit string
}

func NewHealthController(db dbPinger, commit string) *HealthController {
	return &HealthController{db: db, commit: commit}
}

// Health reports whether the API can reach the database and which commit is
// running. The deploy smoke test checks both.
func (h *HealthController) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), healthDBTimeout)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		log.Printf("WARN health: database ping failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "degraded",
			"db":     "unreachable",
			"commit": h.commit,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"db":     "connected",
		"commit": h.commit,
	})
}
