package pages

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"reverse-proxy/models"
	"sort"
	"time"
)

func Homepage(c *gin.Context, namespace string, k8sClientSet *kubernetes.Clientset) {
	pods, err := k8sClientSet.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: "spark-role=driver",
	})

	if err != nil {
		panic(err.Error())
	}

	var sparkApps []models.Application
	for _, pod := range pods.Items {
		var startTime *v1.Time = nil
		var endTime *v1.Time = nil
		var duration *time.Duration = nil
		if len(pod.Status.ContainerStatuses) > 0 {
			containerState := pod.Status.ContainerStatuses[0].State

			if containerState.Running != nil {
				startTime = &containerState.Running.StartedAt
			}

			if containerState.Terminated != nil {
				startTime = &containerState.Terminated.StartedAt
				endTime = &containerState.Terminated.FinishedAt
				_duration := endTime.Time.Sub(containerState.Terminated.StartedAt.Time)
				duration = &_duration
			}

		}

		m := models.Application{
			Id:          pod.Labels["spark-app-selector"],
			Name:        pod.Labels["spark-app-name"],
			Driver:      pod.Name, //pod.Name,
			Status:      fmt.Sprint(pod.Status.Phase),
			StartTime:   fmt.Sprint(startTime),
			EndTime:     fmt.Sprint(endTime),
			Duration:    fmt.Sprint(duration),
			Labels:      pod.Labels,
			Annotations: pod.Annotations,
		}

		sparkApps = append(sparkApps, m)
	}

	sort.Sort(models.StartTimeSorter(sparkApps))
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "Spark Reverse Proxy", "apps": sparkApps})
}
