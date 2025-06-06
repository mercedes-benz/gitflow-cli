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

// Register registers a plugin with the factory.
func (factory *Factory) Register(pluginInstance core.Plugin) {
	factory.RegisterPlugin(pluginInstance)
}

// RegisterFallback registers a plugin as a fallback plugin.
func (factory *Factory) RegisterFallback(pluginInstance core.Plugin) {
	factory.RegisterFallbackPlugin(pluginInstance)
}

// NewPlugin creates and returns a BasePlugin instance with all dependencies injected.
// Plugin implementations can use this method to get a pre-configured BasePlugin.
func (factory *Factory) NewPlugin(config Config) BasePlugin {
	return BasePlugin{
		Config: config,
		Hooks:  factory.Hooks,
	}
}
