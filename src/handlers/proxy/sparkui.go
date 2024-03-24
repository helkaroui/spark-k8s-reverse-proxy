package proxy

import (
	"errors"
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

var sparkUIAppNameURLRegex = regexp.MustCompile("{{\\s*[$]driverIP\\s*}}")
var sparkUIAppNamespaceURLRegex = regexp.MustCompile("{{\\s*[$]appNamespace\\s*}}")
var sparkUIAppIdRegex = regexp.MustCompile(`(?m)\/proxy\/(spark-\w{32})`)
var sparkUIDriverIdRegex = regexp.MustCompile(`(?m)\/proxy\/([^\/]*)`)

func getSparkUIServiceUrl(sparkUIServiceUrlFormat string, driverIP string, appNamespace string) string {
	return sparkUIAppNamespaceURLRegex.ReplaceAllString(sparkUIAppNameURLRegex.ReplaceAllString(sparkUIServiceUrlFormat, driverIP), appNamespace)
}

func flattenIPAddress(ip string) string {
	return strings.ReplaceAll(ip, ".", "-")
}

func parseId(uri string, regex *regexp.Regexp) (string, error) {
	match := regex.FindAllStringSubmatch(uri, 1)
	if len(match) != 0 {
		return match[0][1], nil
	} else {
		return "", errors.New("No Id found in path.")
	}
}

func ServeSparkUI(c *gin.Context, config *handlers.ApiConfig, namespace string, uiRootPath string, driverToSvcMap map[string]string, appIdToDriverMap map[string]string, k8sClientSet *kubernetes.Clientset) {
	path := c.Param("path")
	// remove / prefix if there is any
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	appId, _ := parseId(c.Request.URL.Path, sparkUIAppIdRegex)
	driverId, _ := parseId(c.Request.Referer(), sparkUIDriverIdRegex)

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

	if val, ok := appIdToDriverMap[appId]; ok {
		appName = val
	} else if driverId != appId {
		appIdToDriverMap[appId] = driverId
		appName = driverId
		log.Printf("Map = %s", appIdToDriverMap)
	}

	pod, err := k8sClientSet.CoreV1().Pods(namespace).Get(c, appName, v1.GetOptions{})
	var driverIP = ""

	if val, ok := driverToSvcMap[appName]; ok {
		driverIP = val
	} else if err == nil {
		driverIP = flattenIPAddress(pod.Status.PodIP)
		driverToSvcMap[appName] = driverIP
		log.Printf("Map = %s", driverToSvcMap)
	}

	// get url for the underlying Spark UI Kubernetes service, which is created by spark-on-k8s-operator
	sparkUIServiceUrl := getSparkUIServiceUrl(config.SparkUIServiceUrl, driverIP, config.SparkApplicationNamespace)
	proxyBasePath := ""
	if config.ModifyRedirectUrl {
		proxyBasePath = fmt.Sprintf("%s/%s", uiRootPath, appName)
	}

	proxy, err := newReverseProxy(sparkUIServiceUrl, path, proxyBasePath, appName)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create reverse proxy for application %s: %s", appName, err.Error()))
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
