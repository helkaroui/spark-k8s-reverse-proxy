package proxy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getSparkUIServiceUrl(t *testing.T) {
	assert.Equal(t, "", getSparkUIServiceUrl("", "app1", "ns1"))
	assert.Equal(t,
		"http://%s-ui-svc.%s.svc.cluster.local:4040",
		getSparkUIServiceUrl(
			"http://%s-ui-svc.%s.svc.cluster.local:4040", "app1", "ns1"))
	assert.Equal(t,
		"http://app1-ui-svc.ns1.svc.cluster.local:4040",
		getSparkUIServiceUrl(
			"http://{{$appName}}-ui-svc.{{$appNamespace}}.svc.cluster.local:4040", "app1", "ns1"))
}

func Test_parseId_AppId(t *testing.T) {

	possibleTestCases := [5]string{
		"/proxy/spark-60eb448201a44300bc30f3ac638a9c58",
		"/proxy/spark-60eb448201a44300bc30f3ac638a9c58/",
		"/proxy/spark-60eb448201a44300bc30f3ac638a9c58/jobs",
		"/proxy/spark-60eb448201a44300bc30f3ac638a9c58/jobs/",
		"/proxy/spark-60eb448201a44300bc30f3ac638a9c58/jobs/adzfazef?q=zdz",
	}

	for _, v := range possibleTestCases {
		result, err := parseId(v, sparkUIAppIdRegex)
		assert.NoError(t, err)
		assert.Equal(t, "spark-60eb448201a44300bc30f3ac638a9c58", result)
	}

	errorTestCases := [5]string{
		"/proxy/spark-60eb448201a44300bc30f",
		"/proxy/any-60eb448201a44300bc30f3ac638a9c58/",
		"/proxy/SPARK-60eb448201a44300bc30f3ac638a9c58/jobs",
		"/proxy/Spark-60eb448201a44300bc30f3ac638a9c58/jobs/",
		"/proxy/spaRk-60eb448201a44300bc30f3ac638a9c58/jobs/adzfazef?q=zdz",
	}

	for _, v := range errorTestCases {
		_, err := parseId(v, sparkUIAppIdRegex)
		assert.Error(t, err)
	}
}

func Test_parseId_DriverId(t *testing.T) {

	possibleTestCases := [4]string{
		"/proxy/compute-pi-1-d95wz",
		"/proxy/compute-pi-1-d95wz/",
		"/proxy/compute-pi-1-d95wz/jobs/",
		"/proxy/compute-pi-1-d95wz/jobs?q=efef",
	}

	for _, v := range possibleTestCases {
		result, err := parseId(v, sparkUIDriverIdRegex)
		assert.NoError(t, err)
		assert.Equal(t, "compute-pi-1-d95wz", result)
	}
}
