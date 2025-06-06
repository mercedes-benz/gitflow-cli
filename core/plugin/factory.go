/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"github.com/mercedes-benz/gitflow-cli/core"
)

// Factory is a factory that injects core dependencies into plugin implementations.
type Factory struct {
	// Reference to global hooks
	Hooks *core.HookRegistry
	// Function to register a plugin
	RegisterPlugin func(core.Plugin)
	// Function to register a fallback plugin
	RegisterFallbackPlugin func(core.Plugin)
}

// NewPluginFactory creates a new Factory with the specified dependencies.
func NewPluginFactory() *Factory {
	return &Factory{
		Hooks:                  core.GlobalHooks,
		RegisterPlugin:         core.RegisterPlugin,
		RegisterFallbackPlugin: core.RegisterFallbackPlugin,
	}
}

// CreatePlugin creates a plugin with injected dependencies.
func (factory *Factory) CreatePlugin(
	creator func(factory *Factory) core.Plugin,
) core.Plugin {
	return creator(factory)
}
