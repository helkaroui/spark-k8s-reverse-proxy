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
