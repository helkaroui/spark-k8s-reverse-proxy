package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func HealthCheck(c *gin.Context, startedAt string) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "started_at": startedAt})
}
