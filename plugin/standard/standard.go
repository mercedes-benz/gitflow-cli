/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package standard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mercedes-benz/gitflow-cli/plugin/core"
)

// NewPlugin creates plugin for the standard workflow.
func NewPlugin() core.Plugin {
	plugin := &standardPlugin{}
	core.GlobalHooks.RegisterHook(pluginName, core.ReleaseStartHooks.BeforeReleaseStartHook, plugin.beforeReleaseStart)
	return plugin
}

func init() {
	core.RegisterFallbackPlugin(NewPlugin())
}

// Name of the standard plugin.
const pluginName = "Standard"

// Precondition file pluginName for standard projects.
const preconditionFile = "version.txt"

// Snapshot qualifier for mvn projects.
const snapshotQualifier = "dev"

// standardPlugin is the plugin for the standard workflow.
type standardPlugin struct {
}

func (p *standardPlugin) String() string {
	return pluginName
}

func (p *standardPlugin) SnapshotQualifier() string {
	return snapshotQualifier
}

// RequiredTools returns list of required command line tools.
func (p *standardPlugin) RequiredTools() []string {
	return []string{}
}

// CheckRequiredFile checks if the plugin can be executed in a project directory.
func (p *standardPlugin) CheckRequiredFile(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, preconditionFile))
	return !os.IsNotExist(err)
}

// Version evaluates the current and next version of the standard project.
func (p *standardPlugin) Version(projectPath string, major, minor, incremental bool) (core.Version, core.Version, error) {
	// current and next version of the standard project
	var currentVersion, nextVersion core.Version
	var errMajor, errMinor, errIncremental error

	// read the version from the precondition file
	if bytes, err := os.ReadFile(filepath.Join(projectPath, preconditionFile)); err != nil {
		return core.NoVersion, core.NoVersion, fmt.Errorf("standard version evaluation failed with %v: %v", err, preconditionFile)
	} else {
		if current, err := core.ParseVersion(strings.Trim(string(bytes), "\n\r")); err != nil {
			return core.NoVersion, core.NoVersion, err
		} else {
			currentVersion = current
		}
	}

	// create the next version of the standard project based on the version increment type
	switch {
	case major && !minor && !incremental:
		// create the next major version of the standard project
		nextVersion, errMajor = currentVersion.Next(core.Major)

	case minor && !major && !incremental:
		// create the next minor version of the standard project
		nextVersion, errMinor = currentVersion.Next(core.Minor)

	case incremental && !major && !minor:
		// create the next incremental version of the standard project
		nextVersion, errIncremental = currentVersion.Next(core.Incremental)

	case !major && !minor && !incremental:
		// version increment type not specified, return the current version as next version
		nextVersion = currentVersion

	default:
		return core.NoVersion, core.NoVersion, fmt.Errorf("unsupported version increment type")
	}

	return currentVersion, nextVersion, errors.Join(errMajor, errMinor, errIncremental)
}

// UpdateProjectVersion updates the project's version
func (p *standardPlugin) UpdateProjectVersion(next core.Version) error {
	if err := os.WriteFile(preconditionFile, []byte(next.String()), 0644); err != nil {
		return fmt.Errorf("failed to write in file %v next project version %v", preconditionFile, next.String())
	}

	return nil
}

func (p *standardPlugin) beforeReleaseStart(repo core.Repository) error {
	return nil
}
