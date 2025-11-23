package config

import (
	"fmt"
	"os"

	"dnstester/pkg/types"

	"gopkg.in/yaml.v3"
)

func LoadConfig(filePath string) (*types.Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// perform basic validation on the configuration
func validateConfig(config *types.Config) error {
	if len(config.Domains) == 0 {
		return fmt.Errorf("no domains defined")
	}

	if len(config.Servers) == 0 {
		return fmt.Errorf("no servers defined")
	}

	for i, server := range config.Servers {
		if server.Name == "" {
			return fmt.Errorf("server %d: name is required", i)
		}
		if server.Address == "" {
			return fmt.Errorf("server %d: address is required", i)
		}
		if len(server.Protocols) == 0 {
			return fmt.Errorf("server %d: at least one protocol must be specified", i)
		}

		validProtocols := map[string]bool{
			"udp": true,
			"tcp": true,
			"dot": true,
			"doh": true,
		}

		for _, protocol := range server.Protocols {
			if !validProtocols[protocol] {
				return fmt.Errorf("server %d: invalid protocol '%s'. Must be one of: udp, tcp, dot, doh", i, protocol)
			}
		}
	}

	return nil
}
