/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import (
	"fmt"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

// python-specific command constants
const (
	python = "python"
)

// Fixed configuration for the Python plugin
var pluginConfig = plugin.Config{
	Name: "python",
	VersionFileNames: []string{
		"pyproject.toml",
		"setup.py",
	},
	VersionQualifier: "dev",
	RequiredTools:    []string{python},
}

// pythonPlugin is the struct implementing the Plugin interface.
type pythonPlugin struct {
	plugin.Plugin
}

// Register the Python plugin
func init() {
	pluginFactory := plugin.NewFactory()

	pythonPlugin := &pythonPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	// Register hooks for this plugin
	pythonPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, pythonPlugin.beforeReleaseStart)
	pythonPlugin.RegisterHook(core.HotfixStartHooks.BeforeHotfixStartHook, pythonPlugin.beforeHotfixStart)

	core.RegisterPlugin(pythonPlugin)
}

// ReadVersion reads the version from the Python project.
// If pyproject.toml exists, it will be prioritized and treated as the version file.
// If pyproject.toml does not exist, the version will be read from setup.py.
func (p *pythonPlugin) ReadVersion(_ core.Repository) (core.Version, error) {
	// TODO: Implement the logic, e.g.by using bump-my-version library
	return core.Version{}, fmt.Errorf("Python plugin is not implemented yet")
}

// WriteVersion writes the version to the Python project.
// If pyproject.toml exists, it will be prioritized and treated as the version file.
// If pyproject.toml does not exist, the version will be written to setup.py.
func (p *pythonPlugin) WriteVersion(_ core.Repository, _ core.Version) error {
	// TODO: Implement the logic, e.g.by using bump-my-version library
	return fmt.Errorf("Python plugin is not implemented yet")
}

// beforeReleaseStart ensures a version is set in the Python project file on the development branch
func (p *pythonPlugin) beforeReleaseStart(_ core.Repository) error {
	// TODO: Implement hook logic (e.g. see npm plugin)
	return nil
}

// beforeHotfixStart ensures a version is set in the Python project file on the production branch
func (p *pythonPlugin) beforeHotfixStart(_ core.Repository) error {
	// TODO: Implement hook logic (e.g. see npm plugin)
	return nil
}
