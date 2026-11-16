// Package cfg is internal to the fixture module.
package cfg

// Config is exposed through a public alias.
type Config struct {
	Timeout int
	Deep    Event // reachable only through Config, so not reported again
}

// Opt is exposed through a parameter.
type Opt struct{ N int }

// Result is exposed through a method result.
type Result struct{}

// Base is exposed by embedding.
type Base struct{ Retries int }

// Item is exposed through a type argument.
type Item struct{}

// Hidden is only used inside function bodies, which is fine.
type Hidden struct{}

// Number is exposed as a type-parameter constraint.
type Number interface{ ~int | ~float64 }

// Event is exposed through containers and interfaces.
type Event struct{}

// Named is exposed as an embedded interface.
type Named interface{ Name() string }

// Level is exposed as a term of a type-set constraint.
type Level int
