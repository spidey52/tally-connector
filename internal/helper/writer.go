package helper

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteToFile(filename string, data any) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %s", err.Error())
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %s", err.Error())
	}

	return nil
}
