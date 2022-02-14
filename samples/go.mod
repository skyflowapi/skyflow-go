module github.com/skyflowapi/skyflow-go/samples

go 1.13

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/skyflowapi/skyflow-go/commonutils v0.0.0-20210830070335-73242cbca8cb
	github.com/skyflowapi/skyflow-go/service-account v0.0.0-20210825145958-6ea84a35159f // indirect
	github.com/skyflowapi/skyflow-go/skyflow v1.0.0

)

replace github.com/skyflowapi/skyflow-go/skyflow => ../skyflow

replace github.com/skyflowapi/skyflow-go/commonutils => ../common-utils
