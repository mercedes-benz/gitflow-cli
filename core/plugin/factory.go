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
}

// NewFactory creates a new Factory with the specified dependencies.
func NewFactory() *Factory {
	return &Factory{
		Hooks: core.GlobalHooks,
	}
}

// NewPlugin creates and returns a Plugin instance with all dependencies injected.
// Plugin implementations can use this method to get a pre-configured Plugin.
func (factory *Factory) NewPlugin(config Config) Plugin {
	return Plugin{
		Config: config,
		Hooks:  factory.Hooks,
	}
}

// NewPluginWithExecutor creates a Plugin with an Executor for the given Docker image.
// The executor mode (docker/native) is read from Viper configuration at runtime.
func (factory *Factory) NewPluginWithExecutor(config Config, image string) Plugin {
	return Plugin{
		Config: config,
		Hooks:  factory.Hooks,
		Executor: &Executor{
			PluginName: config.Name,
			Image:      image,
		},
	}
}
