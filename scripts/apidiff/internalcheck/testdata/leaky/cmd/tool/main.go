// Command tool is a main package, which is not public API.
package main

import "example.com/leaky/internal/cfg"

// Exported in package main does not count.
var Exported cfg.Config

func main() {}
