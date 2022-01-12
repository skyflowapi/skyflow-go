module github.com/skyflowapi/skyflow-go/vault-api

go 1.13

require (
	github.com/cristalhq/jwt/v3 v3.1.0
	github.com/cristalhq/jwt/v4 v4.0.0-beta
	github.com/skyflowapi/skyflow-go/errors v0.0.0-20210830070335-73242cbca8cb
)

replace github.com/skyflowapi/skyflow-go/errors => ../errors
