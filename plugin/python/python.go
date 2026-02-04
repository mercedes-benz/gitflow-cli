/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import (
	"fmt"
	"os/exec"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/plugin/python/manager"
)

// pythonPlugin is the struct implementing the Plugin interface for Python projects.
type pythonPlugin struct {
	plugin.Plugin
	manager manager.VersionManager
}

// Configuration for the Python plugin
var pluginConfig = plugin.Config{
	Name: "python",
	// VersionFileName will be set dynamically by core
	VersionFileNames: []string{
		"pyproject.toml",
		"setup.py",
	},
	VersionQualifier: "dev",
	RequiredTools:    []string{}, // Python is optional - we'll check dynamically
}

// Register the Python plugin
func init() {
	pluginFactory := plugin.NewFactory()

	pythonPlugin := &pythonPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	// Register hooks for this plugin
	pythonPlugin.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, pythonPlugin.beforeReleaseStart)
	pythonPlugin.RegisterHook(core.HotfixStartHooks.BeforeHotfixStartHook, pythonPlugin.beforeHotfixStart)

	// Register plugin directly in core
	core.RegisterPlugin(pythonPlugin)
}

// ReadVersion reads the version from the appropriate Python package manager configuration
func (p *pythonPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	var logs = make([]any, 0)
	defer func() { core.Log(logs...) }()

	projectPath := repository.Local()

	// Step 1: Check if Python is installed
	if !isPythonInstalled() {
		logs = append(logs, "Python is not installed - skipping Python package manager detection")
		return core.Version{}, fmt.Errorf("python is not installed")
	}

	logs = append(logs, "Python detected - checking for package managers")

	// Step 2: Detect and initialize the appropriate manager
	detector := manager.NewManagerDetector()
	mgr, err := detector.Detect(projectPath)
	if err != nil {
		logs = append(logs, fmt.Sprintf("No Python package manager detected: %v", err))
		logs = append(logs, "Falling back to default versioning")
		return core.Version{}, fmt.Errorf("no Python package manager detected: %v", err)
	}

	p.manager = mgr
	logs = append(logs, fmt.Sprintf("Using Python package manager: %s", mgr.GetName()))

	// Step 3: Read version from the detected manager
	versionString, err := mgr.GetVersion()
	if err != nil {
		logs = append(logs, err)
		return core.Version{}, fmt.Errorf("failed to read version from %s: %v", mgr.GetName(), err)
	}

	logs = append(logs, fmt.Sprintf("Read version: %s", versionString))

	// Step 4: Parse the version string
	version, err := core.ParseVersion(versionString)
	if err != nil {
		logs = append(logs, err)
		return core.Version{}, fmt.Errorf("failed to parse version: %v", err)
	}

	return version, nil
}

// WriteVersion writes the version to the appropriate Python package manager configuration
func (p *pythonPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	var logs = make([]any, 0)
	defer func() { core.Log(logs...) }()

	projectPath := repository.Local()

	// If manager is not initialized, detect it
	if p.manager == nil {
		// Check Python installation first
		if !isPythonInstalled() {
			logs = append(logs, "Python is not installed - cannot write version")
			return fmt.Errorf("python is not installed")
		}

		detector := manager.NewManagerDetector()
		mgr, err := detector.Detect(projectPath)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to detect Python package manager: %v", err))
			return fmt.Errorf("failed to detect Python package manager: %v", err)
		}
		p.manager = mgr
	}

	logs = append(logs, fmt.Sprintf("Writing version %s using %s", version.String(), p.manager.GetName()))

	// Write version using the detected manager
	if err := p.manager.SetVersion(version.String()); err != nil {
		logs = append(logs, err)
		return fmt.Errorf("failed to write version: %v", err)
	}

	logs = append(logs, "Version written successfully")
	return nil
}

// beforeReleaseStart ensures a version is set on the development branch
func (p *pythonPlugin) beforeReleaseStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Development.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// Check if version is available
	_, err := p.ReadVersion(repository)
	if err == nil {
		// Version exists, nothing to do
		return nil
	}

	// Check if error is due to missing Python or no package manager
	// If so, we should not try to initialize version (let default plugin handle it)
	if !isPythonInstalled() {
		core.Log("Python not installed - skipping Python plugin initialization")
		return nil
	}

	// Python is installed but no package manager detected
	// Try to create initial version if we can detect a manager
	projectPath := repository.Local()
	detector := manager.NewManagerDetector()
	mgr, err := detector.Detect(projectPath)
	if err != nil {
		// No Python package manager detected, skip initialization
		core.Log("No Python package manager detected - skipping initialization")
		return nil
	}

	p.manager = mgr

	// Version doesn't exist, set it to 1.0.0 with qualifier
	initVersion := core.NewVersion("1", "0", "0", p.Config.VersionQualifier)

	if err := p.WriteVersion(repository, initVersion); err != nil {
		return repository.UndoAllChanges(fmt.Errorf("failed to set initial version: %v", err))
	}

	core.Log(fmt.Sprintf("Set initial project version to %s", initVersion.String()))

	if err := repository.CommitChanges("Set initial project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}

// beforeHotfixStart ensures a version is set on the production branch
func (p *pythonPlugin) beforeHotfixStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Production.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// Check if version is available
	_, err := p.ReadVersion(repository)
	if err == nil {
		// Version exists, nothing to do
		return nil
	}

	// Check if error is due to missing Python or no package manager
	if !isPythonInstalled() {
		core.Log("Python not installed - skipping Python plugin initialization")
		return nil
	}

	// Python is installed but no package manager detected
	projectPath := repository.Local()
	detector := manager.NewManagerDetector()
	mgr, err := detector.Detect(projectPath)
	if err != nil {
		// No Python package manager detected, skip initialization
		core.Log("No Python package manager detected - skipping initialization")
		return nil
	}

	p.manager = mgr

	// Version doesn't exist, set it to 1.0.0 (no qualifier for production)
	initVersion := core.NewVersion("1", "0", "0")

	if err := p.WriteVersion(repository, initVersion); err != nil {
		return repository.UndoAllChanges(fmt.Errorf("failed to set initial version: %v", err))
	}

	core.Log(fmt.Sprintf("Set initial project version to %s", initVersion.String()))

	if err := repository.CommitChanges("Set initial project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}

// isPythonInstalled checks if Python (python3 or python) is available
func isPythonInstalled() bool {
	// Try python3 first (preferred)
	if _, err := exec.LookPath("python3"); err == nil {
		return true
	}

	// Fallback to python
	if _, err := exec.LookPath("python"); err == nil {
		return true
	}

	return false
}
