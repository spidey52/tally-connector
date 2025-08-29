package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	Port     int    `json:"port" yaml:"port"`
	DbUrl    string `json:"db_url" yaml:"db_url"`
	TallyUrl string `json:"tally_url" yaml:"tally_url"`
}

type EnvConfig struct {
	Api    ServiceConfig `json:"api" yaml:"api"`
	Loader ServiceConfig `json:"loader" yaml:"loader"`
}

var (
	config     *EnvConfig
	configOnce sync.Once
	configErr  error
)

// LoadEnvConfig reads the YAML file and decodes it into EnvConfig
func LoadEnvConfig() (*EnvConfig, error) {
	file, err := os.Open("env.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to open env.yaml: %w", err)
	}
	defer file.Close()

	var cfg EnvConfig
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode yaml: %w", err)
	}

	return &cfg, nil
}

// GetEnvConfig returns the singleton EnvConfig instance safely
func GetEnvConfig() (EnvConfig, error) {
	configOnce.Do(func() {
		config, configErr = LoadEnvConfig()
	})
	if configErr != nil {
		return EnvConfig{}, configErr
	}
	return *config, nil
}
