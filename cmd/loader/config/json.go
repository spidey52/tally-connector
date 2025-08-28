package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `json:"database"`
	Tally    TallyConfig    `json:"tally"`
}

type DatabaseConfig struct {
	Technology string `json:"technology"`
	Server     string `json:"server"`
	Port       int    `json:"port"`
	SSL        bool   `json:"ssl"`
	Schema     string `json:"schema"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	LoadMethod string `json:"loadmethod"`
}

type TallyConfig struct {
	Definition string `json:"definition"`
	Server     string `json:"server"`
	Port       int    `json:"port"`
	FromDate   string `json:"fromdate"`
	ToDate     string `json:"todate"`
	Sync       string `json:"sync"`
	Frequency  int    `json:"frequency"`
	Company    string `json:"company"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
