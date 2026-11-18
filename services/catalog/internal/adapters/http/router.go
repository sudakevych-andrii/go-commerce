package httpadapter

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ReadinessCheck reports whether a critical dependency (DB, broker, ...) is usable.
type ReadinessCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

const readinessTimeout = 2 * time.Second

func NewRouter(logger *slog.Logger, checks ...ReadinessCheck) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Operational endpoints live outside /api/v1: they are not part of the public contract.
	r.GET("/healthz", healthz)
	r.GET("/readyz", readyz(logger, checks))

	return r
}

// healthz answers "is the process alive?" and must never touch dependencies.
func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyz answers "can the service handle traffic right now?".
func readyz(logger *slog.Logger, checks []ReadinessCheck) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
		defer cancel()

		failed := make([]string, 0)
		for _, rc := range checks {
			if err := rc.Check(ctx); err != nil {
				// Details go to logs only: error text may reveal infrastructure internals.
				logger.Warn("readiness check failed", "check", rc.Name, "error", err)
				failed = append(failed, rc.Name)
			}
		}

		if len(failed) > 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "failed": failed})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
