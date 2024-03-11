package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "log"
	"net/http"
	"strconv"
)

func Applications(c *gin.Context, namespace string, k8sClientSet *kubernetes.Clientset) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err := strconv.Atoi(c.DefaultQuery("size", "10"))

	pods, err := k8sClientSet.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: "spark-role=driver",
	})

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("No driver pods in namespace \"%s\"", namespace))
		panic(err.Error())
	}

	minIndex := (page - 1) * pageSize
	maxIndex := min(page*pageSize, len(pods.Items))

	c.JSON(http.StatusOK, gin.H{"total": len(pods.Items), "page": page, "size": pageSize, "pods": pods.Items[minIndex:maxIndex]})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
