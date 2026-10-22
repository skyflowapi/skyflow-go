module github.com/skyflowapi/skyflow-go/samples/flowvault

go 1.25.0

require github.com/skyflowapi/skyflow-go/flowvault v0.0.0

require (
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/skyflowapi/skyflow-go/common v0.0.0-20261019104456-6115e81fcef9 // indirect
)

// FlowVault isn't published yet; build the samples against the code in this repository.
replace github.com/skyflowapi/skyflow-go/flowvault => ../../flowvault
