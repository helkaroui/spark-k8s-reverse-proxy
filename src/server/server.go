package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	//Init The k8s client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	//kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	//// use the current context in kubeconfig
	//k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	//if err != nil {
	//	panic(err.Error())
	//}
	k8sClientSet, err := kubernetes.NewForConfig(k8sConfig)
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

	var appToSvcMap = make(map[string]string)

	router.GET(fmt.Sprintf("/%s/*path", config.ProxyBaseUri),
		func(context *gin.Context) {
			proxy.ServeSparkUI(context, &apiConfig, config.SparkApplicationNamespace, fmt.Sprintf("/%s", config.ProxyBaseUri), appToSvcMap, k8sClientSet)
		})

	router.Run(fmt.Sprintf(":%d", port))
}
