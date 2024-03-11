package server

type Config struct {
	Port                      int
	SparkApplicationNamespace string
	DriverSvc                 string
	ModifyRedirectUrl         bool
	ProxyBaseUri              string
	TemplatesPath             string
}
