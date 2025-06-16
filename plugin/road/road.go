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
	"regexp"
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

	// Check if the file exists
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		return core.ParseVersion("1.0.0")
	}

	// Read directly from the file with a simple approach
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return core.ParseVersion("1.0.0") // Fallback for safety
	}

	// Search for versionNumber: using regex
	re := regexp.MustCompile(versionKey + `:\s*(.+)`)
	matches := re.FindSubmatch(data)

	if len(matches) >= 2 {
		versionStr := strings.TrimSpace(string(matches[1]))
		return core.ParseVersion(versionStr)
	}

	// Fallback if no version was found
	return core.ParseVersion("1.0.0")
}

// WriteVersion writes the version to the road.yaml file
func (p *roadPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	versionFile := filepath.Join(repository.Local(), p.Config.VersionFileName)

	//// If the file does not exist, create a minimal template
	//if _, err := os.Stat(versionFile); os.IsNotExist(err) {
	//	content := "versionNumber: " + version.String() + "\n"
	//	return os.WriteFile(versionFile, []byte(content), 0644)
	//}

	// Read the content
	data, err := os.ReadFile(versionFile)
	if err != nil {
		// In case of an error, simply create a new file
		//content := "versionNumber: " + version.String() + "\n"
		//return os.WriteFile(versionFile, []byte(content), 0644)
		return fmt.Errorf("road version update failed: %v", err)
	}

	// Replace the version using regex
	re := regexp.MustCompile(versionKey + `:\s*.+`)
	newContent := re.ReplaceAllString(string(data), versionKey+": "+version.String())

	// If no replacement occurred, add the version at the beginning
	if newContent == string(data) {
		newContent = versionKey + ": " + version.String() + "\n" + newContent
	}

	// Write back to the file
	return os.WriteFile(versionFile, []byte(newContent), 0644)
}

// readYamlFile reads and parses a YAML file, returning the parsed data and the raw file content
func (p *roadPlugin) readYamlFile(filePath string) (map[string]interface{}, []byte, error) {
	rawData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %v", p.Config.VersionFileName, err)
	}

	//// Handle empty file case
	//if len(rawData) == 0 {
	//	return make(map[string]interface{}), rawData, nil
	//}

	var dataMap map[string]interface{}
	if err := yaml.Unmarshal(rawData, &dataMap); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML from %s: %v", p.Config.VersionFileName, err)
	}

	//// Initialize map if nil
	//if dataMap == nil {
	//	dataMap = make(map[string]interface{})
	//}

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
