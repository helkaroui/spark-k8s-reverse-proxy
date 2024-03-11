package api

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "log"
	"net/http"
	"strings"
)

func Logs(c *gin.Context, namespace string, k8sClientSet *kubernetes.Clientset) {
	podName := c.Param("podName")
	// remove / prefix if there is any
	if strings.HasPrefix(podName, "/") {
		podName = podName[1:]
	}

	podLogOpts := v1.PodLogOptions{}
	req := k8sClientSet.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(c)

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot fetch logs for driver %s : No such pod", podName))
		panic(err.Error())
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("An error occured when streaming logs for driver %s", podName))
		panic(err.Error())
	}
	logs := buf.String()

	c.String(http.StatusOK, logs)
}
