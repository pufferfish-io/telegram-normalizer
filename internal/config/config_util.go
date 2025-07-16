package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	cfg "telegram-normalizer/internal/config/interfaces"

	"gopkg.in/yaml.v3"
)

var (
	once         sync.Once
	rootInstance map[string]yaml.Node
	instanceErr  error
)

func LoadSection[T cfg.Config]() (T, error) {
	var zero T
	section := zero.SectionName()

	if strings.TrimSpace(section) == "" {
		return zero, fmt.Errorf("incorrect section name: %q", section)
	}

	once.Do(func() {
		data, err := os.ReadFile(DefaultConfigPath)

		if err != nil {
			instanceErr = err
			return
		}

		err = yaml.Unmarshal(data, &rootInstance)
		if err != nil {
			instanceErr = err
		}

		fmt.Println("Configuration loaded")
	})

	if instanceErr != nil {
		return zero, instanceErr
	}

	node, ok := rootInstance[section]
	if !ok {
		return zero, fmt.Errorf("section %q not found in config", section)
	}

	var result T
	err := node.Decode(&result)
	if err != nil {
		return zero, fmt.Errorf("failed to decode section %q: %w", section, err)
	}

	return result, nil
}
