/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"github.com/mercedes-benz/gitflow-cli/core"
)

// Plugin provides a default implementation for common Plugin interface methods.
type Plugin struct {
	Config Config
	Hooks  *core.HookRegistry // Shared hook registry for all plugins
}

// String returns the name of the plugin.
func (p *Plugin) String() string {
	return p.Config.Name
}

// VersionFileName returns the filename containing version information.
func (p *Plugin) VersionFileName() string {
	return p.Config.VersionFileName
}

// SetVersionFileName sets the filename containing version information.
func (p *Plugin) SetVersionFileName(fileName string) {
	p.Config.VersionFileName = fileName
}

// VersionFileNames returns optional list of filenames containing version information.
func (p *Plugin) VersionFileNames() []string {
	return p.Config.VersionFileNames
}

// VersionQualifier returns the qualifier for version strings.
func (p *Plugin) VersionQualifier() string {
	return p.Config.VersionQualifier
}

// RequiredTools returns list of required command line tools.
func (p *Plugin) RequiredTools() []string {
	return p.Config.RequiredTools
}

// RegisterHook is a helper method to register a hook function.
func (p *Plugin) RegisterHook(hookType core.HookType, hookFunction core.HookFunction) {
	if p.Hooks != nil {
		p.Hooks.RegisterHook(p.Config.Name, hookType, hookFunction)
	}
}
