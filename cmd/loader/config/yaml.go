package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TablesConfig struct {
	Master      []Table `json:"master"`
	Transaction []Table `json:"transaction"`
}

type Table struct {
	Name       string   `json:"name"`
	Collection string   `json:"collection"`
	Nature     string   `json:"nature"`
	Fields     []Field  `json:"fields"`
	Fetch      []string `json:"fetch"`
	Filters    []string `json:"filters"`
}

type Field struct {
	Name  string `json:"name"`
	Field string `json:"field"`
	Type  string `json:"type"`
}

func LoadTablesConfig() (*TablesConfig, error) {
	file, err := os.Open("tables.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tablesConfig TablesConfig
	if err := yaml.NewDecoder(file).Decode(&tablesConfig); err != nil {
		return nil, err
	}

	return &tablesConfig, nil
}
