/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package standard

import (
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"os"
	"path/filepath"
	"strings"
)

// Fixed configuration for the standard plugin
var pluginConfig = plugin.Config{
	Name:             "standard",
	VersionFileName:  "version.txt",
	VersionQualifier: "dev",
	RequiredTools:    []string{},
}

// standardPlugin is the plugin for the standard workflow.
type standardPlugin struct {
	plugin.Plugin
}

// Register the standard plugin as a fallback plugin
func init() {
	pluginFactory := plugin.NewFactory()

	// Create plugin with pluginFactory to get hooks and other dependencies
	standardPlugin := &standardPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	// Register hooks
	standardPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, standardPlugin.beforeReleaseStart)
	standardPlugin.RegisterHook(core.HotfixStartHooks.BeforeHotfixStartHook, standardPlugin.beforeHotfixStart)
	standardPlugin.RegisterHook(core.HotfixFinishHooks.AfterMergeIntoDevelopmentHook, standardPlugin.afterMergeIntoDevelopment)

	// Register plugin directly in core, bypassing the pluginFactory
	core.RegisterPlugin(standardPlugin)
	core.RegisterFallbackPlugin(standardPlugin)
}

// ReadVersion reads the current version from the project
func (p *standardPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	projectPath := repository.Local()

	// read the version from the version file
	bytes, err := os.ReadFile(filepath.Join(projectPath, p.Config.VersionFileName))
	if err != nil {
		return core.NoVersion, fmt.Errorf("standard version evaluation failed with %v: %v", err, p.Config.VersionFileName)
	}

	// parse the version string using core.ParseVersion
	versionStr := strings.TrimSpace(string(bytes))
	return core.ParseVersion(versionStr)
}

// WriteVersion writes a new version to the project
func (p *standardPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	projectPath := repository.Local()

	// write the version to the version file
	if err := os.WriteFile(filepath.Join(projectPath, p.Config.VersionFileName), []byte(version.String()), 0644); err != nil {
		return fmt.Errorf("standard version update failed with %v: %v", err, p.Config.VersionFileName)
	}

	return nil
}

func (p *standardPlugin) beforeReleaseStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Development.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// Check if a version file already exists
	versionFilePath := filepath.Join(repository.Local(), p.Config.VersionFileName)
	if _, err := os.Stat(versionFilePath); err == nil {
		return nil
	}

	initVersion := core.NewVersion("1", "0", "0", p.Config.VersionQualifier)
	if err := os.WriteFile(versionFilePath, []byte(initVersion.String()), 0644); err != nil {
		return repository.UndoAllChanges(err)
	}

	if err := repository.AddFile(versionFilePath); err != nil {
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
	versionFilePath := filepath.Join(repository.Local(), p.Config.VersionFileName)
	if _, err := os.Stat(versionFilePath); err == nil {
		return nil
	}

	initVersion := core.NewVersion("1", "0", "0")
	if err := os.WriteFile(versionFilePath, []byte(initVersion.String()), 0644); err != nil {
		return repository.UndoAllChanges(err)
	}

	if err := repository.AddFile(versionFilePath); err != nil {
		return repository.UndoAllChanges(err)
	}

	if err := repository.CommitChanges("Create versions file"); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}

func (p *standardPlugin) afterMergeIntoDevelopment(repository core.Repository) error {

	filesEqual, err := repository.CompareFiles(core.Production.String(), core.Development.String(), p.Config.VersionFileName, p.Config.VersionFileName)

	if err != nil {
		return repository.UndoAllChanges(err)
	}

	// if versions are identical, update the version in the development branch (possible only if hotfix start created initial version)
	if filesEqual {
		if current, err := p.ReadVersion(repository); err != nil {
			return repository.UndoAllChanges(err)
		} else if next, err := current.Next(core.Minor); err != nil {
			return repository.UndoAllChanges(err)
		} else if err := p.WriteVersion(repository, next.AddQualifier(p.VersionQualifier())); err != nil {
			return repository.UndoAllChanges(err)
		}

		if err := repository.CommitChanges("Set next minor project version."); err != nil {
			return repository.UndoAllChanges(err)
		}
	}

	// If different versions, do nothing and proceed normally
	return nil
}
