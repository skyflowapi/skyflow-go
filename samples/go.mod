module github.com/skyflowapi/skyflow-go/samples

go 1.15

require (
	github.com/skyflowapi/skyflow-go/service-account v0.0.0-20210825145958-6ea84a35159f
	github.com/skyflowapi/skyflow-go/skyflow v1.0.0

)

replace github.com/skyflowapi/skyflow-go/skyflow => ../skyflow

replace github.com/skyflowapi/skyflow-go/errors => ../errors
