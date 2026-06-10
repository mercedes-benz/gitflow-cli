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

// TestConfig provides test data for e2e tests.
// Each plugin exports its own TestConfig so that e2e tests can run generically.
type TestConfig struct {
	// Name identifies the test config (used as subtest name)
	Name string
	// PluginName is the plugin's registered name used for executor container lookup.
	// If empty, Name is used.
	PluginName string
	// DockerImage is the container image used for running plugin commands in tests
	DockerImage string
	// VersionQualifier is the qualifier appended to development versions (e.g., "SNAPSHOT", "dev")
	VersionQualifier string
	// Template is the Go template content for the version file
	Template string
	// VersionFileName is the resulting file name (e.g., "pom.xml", "package.json")
	VersionFileName string
	// EmptyFileContent is the content for hook tests (file without version info)
	EmptyFileContent []byte
	// HasBeforeStartHook indicates whether the plugin registers beforeReleaseStart/beforeHotfixStart hooks
	HasBeforeStartHook bool
	// SetupCommands are commands executed inside the container after start (e.g., installing tools)
	SetupCommands [][]string
}
