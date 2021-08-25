module github.com/skyflowapi/skyflow-go/service-account

go 1.15

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/skyflowapi/skyflow-go/errors v0.0.1
)

replace github.com/skyflowapi/skyflow-go/errors => ../errors/
