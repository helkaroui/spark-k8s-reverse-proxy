package pages

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Manifest(c *gin.Context) {
	podName := c.Param("podName")
	// remove / prefix if there is any
	if strings.HasPrefix(podName, "/") {
		podName = podName[1:]
	}

	c.HTML(http.StatusOK, "manifest.tmpl", gin.H{"title": "Driver Manifest", "podName": podName})
}
