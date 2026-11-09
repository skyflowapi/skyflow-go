package a

// Config is declared here before the move.
type Config struct {
	Name  string
	inner int
}

// Get returns the name.
func (c Config) Get() string { return c.Name }

// Set sets the name.
func (c *Config) Set(n string) { c.Name = n }

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
