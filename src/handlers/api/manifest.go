package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	YamlSerializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	_ "log"
	"net/http"
	"strings"
)

func Manifest(c *gin.Context, namespace string, k8sClientSet *kubernetes.Clientset) {
	podName := c.Param("podName")
	// remove / prefix if there is any
	if strings.HasPrefix(podName, "/") {
		podName = podName[1:]
	}

	pod, err := k8sClientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metaV1.GetOptions{})

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot fetch logs for driver %s : No such pod", podName))
		panic(err.Error())
	}

	buf := new(bytes.Buffer)

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("An error occured when streaming logs for driver %s", podName))
		panic(err.Error())
	}

	e := YamlSerializer.NewYAMLSerializer(YamlSerializer.DefaultMetaFactory, nil, nil)
	e.Encode(pod, buf)

	manifest := buf.String()

	c.String(http.StatusOK, manifest)
}
