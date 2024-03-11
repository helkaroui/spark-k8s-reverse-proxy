package pages

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Logs(c *gin.Context) {
	podName := c.Param("podName")
	// remove / prefix if there is any
	if strings.HasPrefix(podName, "/") {
		podName = podName[1:]
	}

	c.HTML(http.StatusOK, "logs.tmpl", gin.H{"title": "Driver Logs", "podName": podName})
}
