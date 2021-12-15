module github.com/skyflowapi/skyflow-go/vault-api

go 1.13

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/skyflowapi/skyflow-go/errors v0.0.0-20210830070335-73242cbca8cb
)

replace github.com/skyflowapi/skyflow-go/errors => ../errors
