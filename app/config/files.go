package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// Read opens and unmarshals the contents of path into a Config configuration
// structure. If any validation errors are encountered, they are returned here
// as well.
func Read(path string) (*Config, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := new(Config)

	err = yaml.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// default to configuration version 1 if not specified.
	if cfg.Version < 1 {
		cfg.Version = 1
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
