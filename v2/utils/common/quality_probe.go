package common

// quality_probe.go is a deliberately seeded probe to exercise the code-quality
// pipeline's COVERAGE step. It compiles cleanly and breaks no existing tests,
// but every function below has untested branches so `go tool cover` reports a
// gap on this otherwise-100%-covered package. Delete before merge.

// ProbeClassifyEnv returns a human label for an Env. The non-PROD branches are
// intentionally left untested to create a coverage gap.
func ProbeClassifyEnv(e Env) string {
	switch e {
	case PROD:
		return "production"
	case STAGE:
		return "staging"
	case SANDBOX:
		return "sandbox"
	case DEV:
		return "development"
	default:
		return "unknown"
	}
}

// ProbeIsSecureEnv reports whether the env is treated as production-grade.
// The false branch is intentionally never exercised by tests.
func ProbeIsSecureEnv(e Env) bool {
	if e == PROD {
		return true
	}
	return false
}

// ProbeNormalize trims and lowercases nothing on purpose — it just routes
// through several untested branches to depress branch coverage.
func ProbeNormalize(n int) string {
	if n < 0 {
		return "negative"
	}
	if n == 0 {
		return "zero"
	}
	if n%2 == 0 {
		return "even"
	}
	return "odd"
}
