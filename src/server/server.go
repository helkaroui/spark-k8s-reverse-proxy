package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"reverse-proxy/handlers"
	"reverse-proxy/handlers/api"
	"reverse-proxy/handlers/pages"
	"reverse-proxy/handlers/proxy"
	"time"
)

func Run(config Config) {
	port := config.Port

	router := gin.Default()
	router.LoadHTMLGlob(config.TemplatesPath)

	apiConfig := handlers.ApiConfig{
		SparkApplicationNamespace: config.SparkApplicationNamespace,
		SparkUIServiceUrl:         config.DriverSvc,
		ModifyRedirectUrl:         config.ModifyRedirectUrl,
	}

	serverStartTime := time.Now()

	k8sClientSet, err := kubernetes.NewForConfig(getK8sConfig())
	if err != nil {
		panic(err.Error())
	}

	router.GET("/",
		func(context *gin.Context) {
			pages.Homepage(context, config.SparkApplicationNamespace, k8sClientSet)
		})

	router.GET("/logs/*podName",
		func(context *gin.Context) {
			pages.Logs(context)
		})

	router.GET("/manifest/*podName",
		func(context *gin.Context) {
			pages.Manifest(context)
		})

	router.GET("/health",
		func(context *gin.Context) {
			api.HealthCheck(context, serverStartTime.Format(time.RFC3339))
		})

	router.GET("/api/applications", func(context *gin.Context) {
		api.Applications(context, config.SparkApplicationNamespace, k8sClientSet)
	})

	router.GET("/api/logs/*podName", func(context *gin.Context) {
		api.Logs(context, config.SparkApplicationNamespace, k8sClientSet)
	})

	router.GET("/api/manifest/*podName", func(context *gin.Context) {
		api.Manifest(context, config.SparkApplicationNamespace, k8sClientSet)
	})

	var driverToSvcMap = make(map[string]string)
	var driverToAppIdMap = make(map[string]string)

	router.GET(fmt.Sprintf("/proxy/*path"),
		func(context *gin.Context) {
			proxy.ServeSparkUI(context, &apiConfig, config.SparkApplicationNamespace, fmt.Sprintf("/%s", config.ProxyBaseUri), driverToSvcMap, driverToAppIdMap, k8sClientSet)
		})

	router.Run(fmt.Sprintf(":%d", port))
}

func getK8sConfig() *rest.Config {
	if os.Getenv("IN_CLUSTER") != "" {
		k8sConfig, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		return k8sConfig
	} else {
		configPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		k8sConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			panic(err.Error())
		}

		return k8sConfig
	}
}
