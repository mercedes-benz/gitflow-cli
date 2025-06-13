/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package npm

import (
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"os/exec"
	"strings"
)

// npm-specific command constant
const (
	npm = "npm"
)

// Fixed configuration for the NPM plugin
var pluginConfig = plugin.Config{
	Name:             "npm",
	VersionFileName:  "package.json",
	VersionQualifier: "dev",
	RequiredTools:    []string{npm},
}

// npmPlugin is the struct implementing the Plugin interface.
type npmPlugin struct {
	plugin.Plugin
}

// Register the NPM plugin
func init() {
	pluginFactory := plugin.NewFactory()

	// Create plugin with pluginFactory to get hooks and other dependencies
	npmPlugin := &npmPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	// Register hooks for this plugin (currently none, but structure is ready for future hooks)
	// Example hook registration would look like this:
	// npmPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, npmPlugin.beforeReleaseStart)

	// Register plugin directly in core, bypassing the pluginFactory
	core.RegisterPlugin(npmPlugin)
}

// ReadVersion reads the version from package.json using npm.
func (p *npmPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	var logs = make([]any, 0)
	// Execute npm command to read the version from package.json
	cmd := exec.Command(npm, "pkg", "get", "version")
	cmd.Dir = repository.Local()

	// log human-readable description of commands
	defer func() { core.Log(logs...) }()

	output, err := cmd.CombinedOutput()
	if err != nil {
		logs = append(logs, cmd, output, err)
		return core.Version{}, fmt.Errorf("failed to read version: %v", err)
	}

	logs = append(logs, cmd, output)
	// Clean the version string
	versionString := strings.TrimSpace(string(output))
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
	var err error
	var cmd *exec.Cmd
	var output []byte

	// Execute npm command to write the version to package.json
	cmd = exec.Command(npm, "version", version.String(), "--no-git-tag-version")
	cmd.Dir = repository.Local()

	// log human-readable description of the npm command
	defer func() { core.Log(cmd, output, err) }()

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write version: %v: %s", err, output)
	}

	return nil
}
