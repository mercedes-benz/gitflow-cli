/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package road

import (
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// road-specific constants
const (
	versionKey = "versionNumber"
)

// Fixed configuration for the Road plugin
var pluginConfig = plugin.Config{
	Name:             "road",
	VersionFileName:  "road.yaml",
	VersionQualifier: "dev",
	RequiredTools:    []string{},
}

// roadPlugin is the struct implementing the Plugin interface.
type roadPlugin struct {
	plugin.Plugin
}

// Register the Road plugin
func init() {
	pluginFactory := plugin.NewFactory()

	// Create plugin with pluginFactory to get hooks and other dependencies
	roadPlugin := &roadPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	// Register hooks for this plugin (currently none, but structure is ready for future hooks)
	// Example hook registration would look like this:
	// roadPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, roadPlugin.beforeReleaseStart)

	// Register plugin directly in core
	core.RegisterPlugin(roadPlugin)
}

// ReadVersion reads the version from road.yaml file
func (p *roadPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	versionFile := filepath.Join(repository.Local(), p.Config.VersionFileName)

	// Read and parse the YAML file
	dataMap, _, err := p.readYamlFile(versionFile)
	if err != nil {
		return core.NoVersion, fmt.Errorf("road version evaluation failed: %v", err)
	}

	// Extract version from YAML rawData
	versionStr, err := p.extractVersion(dataMap)
	if err != nil {
		return core.NoVersion, fmt.Errorf("road version evaluation failed: %v", err)
	}

	// parse the version string using core.ParseVersion
	return core.ParseVersion(versionStr)

}

// WriteVersion writes the version to the road.yaml file
func (p *roadPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	versionFile := filepath.Join(repository.Local(), p.Config.VersionFileName)

	// Read and parse the YAML file
	dataMap, _, err := p.readYamlFile(versionFile)
	if err != nil {
		return fmt.Errorf("road version update failed: %v", err)
	}

	// Update the version in the YAML data
	dataMap[versionKey] = version.String()

	// Write the updated YAML back to the file
	if err := p.writeYamlFile(versionFile, dataMap); err != nil {
		return fmt.Errorf("road version update failed: %v", err)
	}

	return nil
}

// readYamlFile reads and parses a YAML file, returning the parsed data and the raw file content
func (p *roadPlugin) readYamlFile(filePath string) (map[string]interface{}, []byte, error) {
	rawData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %v", p.Config.VersionFileName, err)
	}

	var dataMap map[string]interface{}
	if err := yaml.Unmarshal(rawData, &dataMap); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML from %s: %v", p.Config.VersionFileName, err)
	}

	return dataMap, rawData, nil
}

// extractVersion extracts the version string from road.yaml file
func (p *roadPlugin) extractVersion(dataMap map[string]interface{}) (string, error) {
	versionInterface, found := dataMap[versionKey]
	if !found || versionInterface == nil {
		return "", fmt.Errorf("version key not found")
	}

	version, ok := versionInterface.(string)
	if !ok {
		return "", fmt.Errorf("version is not a string")
	}

	return strings.TrimSpace(version), nil
}

// writeYamlFile writes YAML data to a file while preserving the original format
func (p *roadPlugin) writeYamlFile(filePath string, data map[string]interface{}) error {
	// Read the current file content to get the structure
	rawData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", p.Config.VersionFileName, err)
	}

	rawDataString := string(rawData)

	// Get the version value we want to set
	version := data[versionKey].(string)

	// Simple string replacement approach to find and replace version line
	// A regular expression could be more robust but this works well for this use case
	replaced := false
	lines := strings.Split(rawDataString, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), versionKey+":") {
			// Replace only the value, keep indentation intact
			indent := strings.Index(line, versionKey)
			lines[i] = strings.Repeat(" ", indent) + versionKey + ": " + version
			replaced = true
			break
		}
	}

	// If version wasn't found, add it at the beginning
	if !replaced {
		lines = append([]string{versionKey + ": " + version}, lines...)
	}

	// Write the updated content back to file
	updatedContent := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write to %s: %v", p.Config.VersionFileName, err)
	}

	return nil
}
