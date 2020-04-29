package differer

import (
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultTimeout = 10 * time.Second

// Config stores the differer engine configuration.
type Config struct {
	Runners map[string]string `yaml:"runners"`
	Timeout time.Duration     `yaml:"timeout"`
}

// ReadConfig reads the configuration into a config object.
// Timeout defaults to 10 seconds.
func ReadConfig(r io.Reader) (*Config, error) {
	var c Config
	if err := yaml.NewDecoder(r).Decode(&c); err != nil {
		return nil, fmt.Errorf("could not decode: %w", err)
	}

	if c.Timeout.Seconds() == 0 {
		c.Timeout = defaultTimeout
	}

	return &c, nil
}
