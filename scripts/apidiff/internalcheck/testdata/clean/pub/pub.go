// Package pub uses internal code only in bodies and constant values.
package pub

import "example.com/clean/internal/cfg"

// DefaultTimeout comes from an internal constant.
const DefaultTimeout = cfg.Timeout

// Run uses an internal type in its body.
func Run() int { return cfg.Config{N: 1}.N }

// Check returns a predeclared type.
func Check() error { return nil }
