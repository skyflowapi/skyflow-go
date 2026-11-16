// Package pub is the fixture's public API, which leaks internal types.
package pub

import "example.com/leaky/internal/cfg"

// Config re-exports an internal type.
type Config = cfg.Config

// New takes an internal type.
func New(o cfg.Opt) *Client { _ = cfg.Hidden{}; return &Client{} }

// Client embeds an internal type and has a method returning one.
type Client struct {
	cfg.Base
	name string
}

// Do returns an internal type.
func (c *Client) Do() cfg.Result { return cfg.Result{} }

// List is generic over an internal type.
type List[T any] struct{ Items []T }

// Items exposes an internal type as a type argument.
var Items List[cfg.Item]

// helper is unexported, so its internal type is fine.
func helper() cfg.Hidden { return cfg.Hidden{} }

// Max is constrained by an internal interface.
func Max[T cfg.Number](a, b T) T { return a }

// Events exposes internal types inside containers.
var (
	ByName  map[string]cfg.Event
	Stream  chan cfg.Event
	Recent  [2]cfg.Event
	Handler interface{ Handle(cfg.Event) }
)

// Entity embeds an internal interface.
type Entity interface{ cfg.Named }

// Severity is a type set with an internal term.
type Severity interface{ cfg.Level }

// Mixed is a union with an internal term.
type Mixed interface{ ~string | cfg.Level }
