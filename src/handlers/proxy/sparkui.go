package proxy

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"reverse-proxy/handlers"
	"strings"
)

var sparkUIAppNameURLRegex = regexp.MustCompile("{{\\s*[$]appName\\s*}}")
var sparkUIAppNamespaceURLRegex = regexp.MustCompile("{{\\s*[$]appNamespace\\s*}}")

func getSparkUIServiceUrl(sparkUIServiceUrlFormat string, appName string, appNamespace string) string {
	return sparkUIAppNamespaceURLRegex.ReplaceAllString(sparkUIAppNameURLRegex.ReplaceAllString(sparkUIServiceUrlFormat, appName), appNamespace)
}

func ServeSparkUI(c *gin.Context, config *handlers.ApiConfig, namespace string, uiRootPath string, appToSvcMap map[string]string, k8sClientSet *kubernetes.Clientset) {
	path := c.Param("path")
	// remove / prefix if there is any
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	// get application name
	appName := ""
	index := strings.Index(path, "/")
	if index <= 0 {
		appName = path
		path = ""
	} else {
		appName = path[0:index]
		path = path[index+1:]
	}

	pods, err := k8sClientSet.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("spark-app-selector=%s,spark-role=driver", appName),
	})

	var driverSvc = ""

	if val, ok := appToSvcMap[appName]; ok {
		driverSvc = val
	} else if len(pods.Items) > 0 {
		driverSvc = pods.Items[0].Name
		appToSvcMap[appName] = driverSvc
		log.Printf("Map = %s", appToSvcMap)
	}

	// get url for the underlying Spark UI Kubernetes service, which is created by spark-on-k8s-operator
	sparkUIServiceUrl := getSparkUIServiceUrl(config.SparkUIServiceUrl, driverSvc, config.SparkApplicationNamespace)
	proxyBasePath := ""
	if config.ModifyRedirectUrl {
		proxyBasePath = fmt.Sprintf("%s/%s", uiRootPath, driverSvc)
	}
	proxy, err := newReverseProxy(sparkUIServiceUrl, path, proxyBasePath, driverSvc)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create reverse proxy for application %s: %s", driverSvc, err.Error()))
		return
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func newReverseProxy(sparkUIServiceUrl string, targetPath string, proxyBasePath string, appName string) (*httputil.ReverseProxy, error) {
	log.Printf("Creating revers proxy for Spark UI service url %s", sparkUIServiceUrl)
	targetUrl := sparkUIServiceUrl
	if targetPath != "" {
		if !strings.HasPrefix(targetPath, "/") {
			targetPath = "/" + targetPath
		}
		targetUrl += targetPath
	}
	url, err := url.Parse(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target Spark UI url %s: %s", targetUrl, err.Error())
	}
	director := func(req *http.Request) {
		url.RawQuery = req.URL.RawQuery
		url.RawFragment = req.URL.RawFragment
		req.URL = url

		//req.Header.Add("X-Forwarded-Context", appName)
		log.Printf("Reverse proxy: serving backend url %s for originally requested url %s", url, req.URL)
		log.Printf("Headers X-Forwarded-Context=%s", req.Header.Get("X-Forwarded-Context"))
	}

	modifyResponse := func(resp *http.Response) error {
		//resp.Header.Add("X-Forwarded-Context", appName)
		if proxyBasePath != "" && resp.StatusCode == http.StatusFound {
			// Append the proxy base path before the redirect path.
			// Also modify redirect url to only contain path and not contain host name,
			// so redirect will retain the original requested host name.
			headerName := "Location"
			locationHeaderValues := resp.Header[headerName]
			if len(locationHeaderValues) > 0 {
				newValues := make([]string, 0, len(locationHeaderValues))
				for _, oldHeaderValue := range locationHeaderValues {
					parsedUrl, err := url.Parse(oldHeaderValue)
					if err != nil {
						log.Printf("Reverse proxy: invalid response header value %s: %s (backend url %s): %s", headerName, oldHeaderValue, url, err.Error())
						newValues = append(newValues, oldHeaderValue)
					} else {
						parsedUrl.Scheme = ""
						parsedUrl.Host = ""
						newPath := parsedUrl.Path
						if !strings.HasPrefix(newPath, "/") {
							newPath = "/" + newPath
						}
						parsedUrl.Path = proxyBasePath + newPath
						newHeaderValue := parsedUrl.String()
						log.Printf("Reverse proxy: modifying response header %s from %s to %s (backend url %s)", headerName, oldHeaderValue, newHeaderValue, url)
						newValues = append(newValues, newHeaderValue)
					}
				}
				resp.Header[headerName] = newValues
			}
		}
		return nil
	}
	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
	}, nil
}
