/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"github.com/spf13/viper"
)

// Config contains configuration values for plugin-specific behavior.
// This structure allows plugin constants to be parameterized instead of
// being hardcoded in each plugin.
type Config struct {
	// Name of the plugin for display and registration purposes
	Name string
	// File name that contains version information
	VersionFileName string
	// Qualifier for SNAPSHOT versions
	VersionQualifier string
	// Required external tools
	RequiredTools []string
}

// LoadPluginConfig loads the configuration for a plugin from the Viper configuration.
// If no specific configuration is found, default values will be returned.
func LoadPluginConfig(pluginName string, defaults Config) Config {
	config := defaults

	// Configuration path for plugins
	configPath := "plugins." + pluginName

	// Try to load configuration values if they exist
	if viper.IsSet(configPath + ".versionFile") {
		config.VersionFileName = viper.GetString(configPath + ".versionFile")
	}

	if viper.IsSet(configPath + ".versionQualifier") {
		config.VersionQualifier = viper.GetString(configPath + ".versionQualifier")
	}

	if viper.IsSet(configPath + ".requiredTools") {
		config.RequiredTools = viper.GetStringSlice(configPath + ".requiredTools")
	}

	return config
}
