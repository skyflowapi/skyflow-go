package impl

// Config moved here.
type Config struct {
	Name  string
	inner int
}

// Get returns the name.
func (c Config) Get() string { return c.Name }

// Set sets the name.
func (c *Config) Set(n string) { c.Name = n }
