package properties

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const propertiesFile = "properties.json"

// Properties holds user preferences and active CLI state.
type Properties struct {
	App string `json:"app,omitempty"`
}

// Load reads properties from the config directory.
// Returns empty Properties (not an error) if the file does not exist.
func Load(configDir string) (*Properties, error) {
	path := filepath.Join(configDir, propertiesFile)
	data, err := os.ReadFile(path) //nolint:gosec // path is constructed from configDir + constant filename
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Properties{}, nil
		}
		return nil, fmt.Errorf("failed to read properties: %w", err)
	}

	var props Properties
	if err := json.Unmarshal(data, &props); err != nil {
		return nil, fmt.Errorf("failed to parse properties: %w", err)
	}

	return &props, nil
}

// Save writes properties to the config directory.
func Save(configDir string, props *Properties) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(props, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal properties: %w", err)
	}

	path := filepath.Join(configDir, propertiesFile)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write properties: %w", err)
	}

	return nil
}

// Clear removes the properties file.
func Clear(configDir string) error {
	path := filepath.Join(configDir, propertiesFile)
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
