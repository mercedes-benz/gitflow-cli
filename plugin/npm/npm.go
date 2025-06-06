/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package npm

import (
	"bytes"
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"os/exec"
	"strings"
)

// Default configuration for the NPM plugin
var defaultConfig = plugin.Config{
	Name:             "NPM",
	VersionFileName:  "package.json",
	VersionQualifier: "dev",
	RequiredTools:    []string{"npm"},
}

// npmPlugin is the struct implementing the Plugin interface.
type npmPlugin struct {
	plugin.BasePlugin
}

// NewPlugin creates a plugin for the NPM build tool.
func NewPlugin(factory *plugin.Factory) core.Plugin {
	// Load configurable values from configuration
	config := plugin.LoadPluginConfig(defaultConfig.Name, defaultConfig)

	npmPlugin := &npmPlugin{
		BasePlugin: plugin.BasePlugin{
			Config: config,
			Hooks:  factory.Hooks,
		},
	}

	// Register hooks for this plugin (currently none, but structure is ready for future hooks)
	// Example hook registration would look like this:
	// npmPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, npmPlugin.beforeReleaseStart)

	return npmPlugin
}

// Register the NPM plugin
func init() {
	factory := plugin.NewPluginFactory()
	pluginInstance := factory.CreatePlugin(func(f *plugin.Factory) core.Plugin {
		return NewPlugin(f)
	})
	factory.RegisterPlugin(pluginInstance)
}

// ReadVersion reads the version from package.json using npm.
func (p *npmPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	// Execute npm command to read the version from package.json
	cmd := exec.Command("npm", "pkg", "get", "version")
	cmd.Dir = repository.Local()

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return core.Version{}, fmt.Errorf("failed to read version: %v", err)
	}

	// Clean the version string
	versionString := strings.TrimSpace(stdout.String())
	// Remove surrounding quotes from the npm output
	versionString = strings.Trim(versionString, "\"")

	// Parse the version string
	version, err := core.ParseVersion(versionString)
	if err != nil {
		return core.Version{}, fmt.Errorf("failed to parse version: %v", err)
	}

	return version, nil
}

// WriteVersion writes the version to package.json using npm.
func (p *npmPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	// Execute npm command to write the version to package.json
	cmd := exec.Command("npm", "version", version.String(), "--no-git-tag-version")
	cmd.Dir = repository.Local()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write version: %v", err)
	}

	return nil
}
