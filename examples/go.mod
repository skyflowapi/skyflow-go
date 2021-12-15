module github.com/skyflowapi/skyflow-go/examples

go 1.15

require (
	github.com/skyflowapi/skyflow-go/errors v0.0.0-20210830070335-73242cbca8cb
	github.com/skyflowapi/skyflow-go/service-account v0.0.0-20210825145958-6ea84a35159f
	github.com/skyflowapi/skyflow-go/vault-api v0.0.0-00010101000000-000000000000
)

replace github.com/skyflowapi/skyflow-go/vault-api => ../vault-api
