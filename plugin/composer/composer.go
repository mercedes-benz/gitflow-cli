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

	// Register hooks for this plugin
	composerPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, composerPlugin.beforeReleaseStart)
	composerPlugin.RegisterHook(core.HotfixStartHooks.BeforeHotfixStartHook, composerPlugin.beforeHotfixStart)

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

// beforeReleaseStart ensures a version is set in the composer.json file on the development branch
func (p *composerPlugin) beforeReleaseStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Development.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// Check if version is available in composer.json
	_, err := p.ReadVersion(repository)
	if err == nil {
		// Version exists, nothing to do
		return nil
	}

	// Version doesn't exist, set it to 1.0.0 with qualifier
	initVersion := core.NewVersion("1", "0", "0", p.Config.VersionQualifier)

	// Set the version using composer CLI
	cmd := exec.Command(composer, "config", "version", initVersion.String(), "--no-ansi")
	cmd.Dir = repository.Local()

	output, err := cmd.CombinedOutput()
	if err != nil {
		core.Log(cmd, output, err)
		return repository.UndoAllChanges(fmt.Errorf("failed to set initial version: %v", err))
	}

	core.Log(cmd, output)

	if err := repository.CommitChanges("Set initial project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}

// beforeHotfixStart ensures a version is set in the composer.json file on the production branch
func (p *composerPlugin) beforeHotfixStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Production.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// Check if version is available in composer.json
	_, err := p.ReadVersion(repository)
	if err == nil {
		// Version exists, nothing to do
		return nil
	}

	// Version doesn't exist, set it to 1.0.0 (no qualifier for production)
	initVersion := core.NewVersion("1", "0", "0")

	// Set the version using composer CLI
	cmd := exec.Command(composer, "config", "version", initVersion.String(), "--no-ansi")
	cmd.Dir = repository.Local()

	output, err := cmd.CombinedOutput()
	if err != nil {
		core.Log(cmd, output, err)
		return repository.UndoAllChanges(fmt.Errorf("failed to set initial version: %v", err))
	}

	core.Log(cmd, output)

	if err := repository.CommitChanges("Set initial project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}
