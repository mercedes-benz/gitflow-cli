/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package composer

import (
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"os/exec"
	"strings"
)

// composer-specific command constants
const (
	composer = "composer"
)

// Fixed configuration for the Composer plugin
var pluginConfig = plugin.Config{
	Name:             "composer",
	VersionFileName:  "composer.json",
	VersionQualifier: "dev",
	RequiredTools:    []string{composer},
}

// composerPlugin is the struct implementing the Plugin interface.
type composerPlugin struct {
	plugin.Plugin
}

// Register the Composer plugin
func init() {
	pluginFactory := plugin.NewFactory()

	composerPlugin := &composerPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	// Register hooks for this plugin (currently none, but structure is ready for future hooks)
	// Example hook registration would look like this:
	// composerPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, composerPlugin.beforeReleaseStart)

	core.RegisterPlugin(composerPlugin)
}

// ReadVersion reads the version from composer.json using composer.
func (p *composerPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	var logs = make([]any, 0)
	// Execute composer command to read the version from composer.json
	cmd := exec.Command(composer, "config", "version", "--no-ansi")
	cmd.Dir = repository.Local()

	// log human-readable description of commands
	defer func() { core.Log(logs...) }()

	output, err := cmd.CombinedOutput()
	if err != nil {
		logs = append(logs, cmd, output, err)
		return core.Version{}, fmt.Errorf("failed to read version from composer.json: %v", err)
	}

	logs = append(logs, cmd, output)
	// Clean the version string
	versionString := strings.TrimSpace(string(output))

	// Parse the version string
	version, err := core.ParseVersion(versionString)
	if err != nil {
		return core.Version{}, fmt.Errorf("failed to parse version: %v", err)
	}

	return version, nil
}

// WriteVersion writes the version to composer.json using composer.
func (p *composerPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	var err error
	var cmd *exec.Cmd
	var output []byte

	// Execute composer command to write the version to composer.json
	cmd = exec.Command(composer, "config", "version", version.String(), "--no-ansi")
	cmd.Dir = repository.Local()

	// log human-readable description of the composer command
	defer func() { core.Log(cmd, output, err) }()

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write version to composer.json: %v: %s", err, output)
	}

	return nil
}
