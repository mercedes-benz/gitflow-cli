/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package road

import (
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// road-specific constants
const (
	versionKey = "versionNumber"
)

var versionRegex = regexp.MustCompile(`(?m)^(` + versionKey + `\s*:)(\s*)(['"]?)(.+?)(['"]?)\s*$`)

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

	// Read directly from the file
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return core.Version{}, fmt.Errorf("failed to read road version file: %v", err)
	}

	// Check for multiple version entries
	allMatches := versionRegex.FindAllSubmatch(data, -1)
	if len(allMatches) > 1 {
		return core.Version{}, fmt.Errorf("multiple version entries found in road.yaml file")
	}

	// Get the first (and should be only) match
	matches := versionRegex.FindSubmatch(data)

	// The version is in the fourth group (index 4)
	if len(matches) >= 5 {
		versionStr := strings.TrimSpace(string(matches[4]))
		return core.ParseVersion(versionStr)
	}

	// No version found in file
	return core.Version{}, fmt.Errorf("no version found in road.yaml file")
}

// WriteVersion writes the version to the road.yaml file
func (p *roadPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	versionFile := filepath.Join(repository.Local(), p.Config.VersionFileName)

	// Read the content
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return fmt.Errorf("road version update failed: %v", err)
	}

	// When replacing, we use exactly one space after the colon and keep the original quotation marks (groups 3 and 5)
	newContent := versionRegex.ReplaceAllString(string(data), "${1} ${3}"+version.String()+"${5}")

	// If no replacement occurred, return an error
	if newContent == string(data) {
		return fmt.Errorf("version key not found in road.yaml file")
	}

	// Write back to the file
	return os.WriteFile(versionFile, []byte(newContent), 0644)
}
