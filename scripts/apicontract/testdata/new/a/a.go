package a

import "example.com/m/internal/impl"

// Config keeps the old name for the moved type.
type Config = impl.Config

// New builds a Config.
func New(name string, opts ...int) (*Config, error) { return &Config{Name: name}, nil }

// Doer is implemented by callers.
type Doer interface{ Do(Config) error }

// Msg is an exported message.
const Msg = "hello"

// Level is a named string.
type Level string

// Alias is an alias of a basic type.
type Alias = string

// Default is a package var.
var Default Config

// Box is generic.
type Box[T any] struct{ V T }

// Sealed can't be implemented outside the module.
type Sealed interface {
	Public() int
	private()
}

// Number is a type-set constraint.
type Number interface{ ~int | ~float64 }
