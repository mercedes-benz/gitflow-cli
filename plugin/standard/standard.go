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
	core.GlobalHooks.RegisterHook(pluginName, core.HotfixStartHooks.BeforeHotfixStartHook, plugin.beforeHotfixStart)
	core.GlobalHooks.RegisterHook(pluginName, core.HotfixFinishHooks.AfterMergeIntoDevelopmentHook, plugin.afterMergeIntoDevelopment)
	return plugin
}

func init() {
	core.RegisterFallbackPlugin(NewPlugin())
}

const pluginName = "Standard"

const versionFile = "version.txt"

const versionQualifier = "dev"

// standardPlugin is the plugin for the standard workflow.
type standardPlugin struct {
}

func (p *standardPlugin) String() string {
	return pluginName
}

func (p *standardPlugin) VersionFile() string {
	return versionFile
}

func (p *standardPlugin) VersionQualifier() string {
	return versionQualifier
}

// RequiredTools returns list of required command line tools.
func (p *standardPlugin) RequiredTools() []string {
	return []string{}
}

// CheckVersionFile checks if the plugin can be executed in a project directory.
func (p *standardPlugin) CheckVersionFile(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, versionFile))
	return !os.IsNotExist(err)
}

// Version evaluates the current and next version of the standard project.
func (p *standardPlugin) Version(projectPath string, major, minor, incremental bool) (core.Version, core.Version, error) {
	// current and next version of the standard project
	var currentVersion, nextVersion core.Version
	var errMajor, errMinor, errIncremental error

	// read the version from the version file
	if bytes, err := os.ReadFile(filepath.Join(projectPath, versionFile)); err != nil {
		return core.NoVersion, core.NoVersion, fmt.Errorf("standard version evaluation failed with %v: %v", err, versionFile)
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
	if err := os.WriteFile(versionFile, []byte(next.String()), 0644); err != nil {
		return fmt.Errorf("failed to write in file %v next project version %v", versionFile, next.String())
	}

	return nil
}

func (p *standardPlugin) beforeReleaseStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Development.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// Check if a version file already exists
	versionFilePath := filepath.Join(repository.Local(), versionFile)
	if _, err := os.Stat(versionFilePath); err == nil {
		return nil
	}

	initVersion := core.NewVersion("1", "0", "0", versionQualifier)
	if err := os.WriteFile(versionFile, []byte(initVersion.String()), 0644); err != nil {
		return repository.UndoAllChanges(err)
	}

	if err := repository.AddFile(versionFile); err != nil {
		return repository.UndoAllChanges(err)
	}

	if err := repository.CommitChanges("Create versions file"); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}

func (p *standardPlugin) beforeHotfixStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Production.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// Check if a version file already exists
	versionFilePath := filepath.Join(repository.Local(), versionFile)
	if _, err := os.Stat(versionFilePath); err == nil {
		return nil
	}

	initVersion := core.NewVersion("1", "0", "0")
	if err := os.WriteFile(versionFile, []byte(initVersion.String()), 0644); err != nil {
		return repository.UndoAllChanges(err)
	}

	if err := repository.AddFile(versionFile); err != nil {
		return repository.UndoAllChanges(err)
	}

	if err := repository.CommitChanges("Create versions file"); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}

func (p *standardPlugin) afterMergeIntoDevelopment(repository core.Repository) error {

	filesEqual, err := repository.CompareFiles(core.Production.String(), core.Development.String(), versionFile, versionFile)

	if err != nil {
		return repository.UndoAllChanges(err)
	}

	// if versions are identical, update the version in the development branch (possible only if hotfix start created initil version)
	if filesEqual {
		if _, next, err := p.Version(repository.Local(), false, true, false); err != nil {
			return repository.UndoAllChanges(err)
		} else if err := p.UpdateProjectVersion(next.AddQualifier(p.VersionQualifier())); err != nil {
			return repository.UndoAllChanges(err)
		}

		if err := repository.CommitChanges("Set next minor project version."); err != nil {
			return repository.UndoAllChanges(err)
		}
	}

	// If different versions, do nothing and proceed normally
	return nil
}
