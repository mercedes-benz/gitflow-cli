/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

// Config contains configuration values for plugin-specific behavior.
type Config struct {
	// Name of the plugin for display and registration purposes
	Name string
	// File name that contains version information
	VersionFileName string
	// Optional list of file names that contain version information (alternative to VersionFileName)
	VersionFileNames []string
	// Qualifier for SNAPSHOT versions
	VersionQualifier string
	// Required external tools
	RequiredTools []string
}
