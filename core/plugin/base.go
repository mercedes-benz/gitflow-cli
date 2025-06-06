/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"github.com/mercedes-benz/gitflow-cli/core"
)

// BasePlugin provides a default implementation for common Plugin interface methods.
// Plugin implementations can embed this struct to avoid code duplication.
type BasePlugin struct {
	Config Config
	Hooks  *core.HookRegistry // Shared hook registry for all plugins
}

// String returns the name of the plugin.
func (p *BasePlugin) String() string {
	return p.Config.Name
}

// VersionFileName returns the filename containing version information.
func (p *BasePlugin) VersionFileName() string {
	return p.Config.VersionFileName
}

// VersionQualifier returns the qualifier for version strings.
func (p *BasePlugin) VersionQualifier() string {
	return p.Config.VersionQualifier
}

// RequiredTools returns list of required command line tools.
func (p *BasePlugin) RequiredTools() []string {
	return p.Config.RequiredTools
}

// RegisterHook is a helper method to register a hook function.
func (p *BasePlugin) RegisterHook(hookType core.HookType, hookFunction core.HookFunction) {
	if p.Hooks != nil {
		p.Hooks.RegisterHook(p.Config.Name, hookType, hookFunction)
	}
}
