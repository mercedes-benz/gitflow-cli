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

// Create plugin for the standard workflow.
func NewPlugIn() core.PlugIn {
	return &standardPlugIn{}
}

// Name of the standard plugin.
const name = "Standard"

// Precondition file name for standard projects.
const preconditionFile = "version.txt"

// StandardPlugIn is the plugin for the standard workflow.
type standardPlugIn struct {
}

// Check if the plugin can be executed in a project directory.
func (p *standardPlugIn) Check(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, preconditionFile))
	return !os.IsNotExist(err)
}

// Version evaluates the current and next version of the standard project.
func (p *standardPlugIn) Version(projectPath string, major, minor, incremental bool) (core.Version, core.Version, error) {
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

// Start command of the standard workflow.
func (p *standardPlugIn) Start(branch core.Branch, projectPath string, args ...any) error {
	var start core.StartCallback

	// select suitable business logic for the branch
	switch branch {
	case core.Release:
		// release business logic
		start = func(repo core.Repository, args ...any) error {
			return p.releaseStart(repo, args[0].(bool), args[1].(bool))
		}

	case core.Hotfix:
		// hotfix business logic
		start = func(repo core.Repository, _ ...any) error {
			return p.hotfixStart(repo)
		}

	default:
		return fmt.Errorf("unsupported branch: %v", branch)
	}

	// start the workflow with the selected business logic
	return core.StartWorkflow(start, []string{core.Git}, branch, name, projectPath, args...)
}

// Finish command of the standard workflow.
func (p *standardPlugIn) Finish(branch core.Branch, projectPath string) error {
	// select suitable business logic for the branch
	switch branch {
	case core.Release:
		// finish the workflow with the selected release business logic
		return core.FinishWorkflow(p.releaseFinish, []string{core.Git}, branch, name, projectPath)

	case core.Hotfix:
		// finish the workflow with the selected hotfix business logic
		return core.FinishWorkflow(p.hotfixFinish, []string{core.Git}, branch, name, projectPath)

	default:
		return fmt.Errorf("unsupported branch: %v", branch)
	}
}

// Register plugin for the standard workflow.
func init() {
	core.Register(NewPlugIn())
}

// Run the release start command for the standard workflow.
func (p *standardPlugIn) releaseStart(repo core.Repository, major, minor bool) error {
	// read out the current and next project version ${major}.${minor}.${increment}-${qualifier}
	_, next, err := p.Version(repo.Local(), major, minor, false)

	if err != nil {
		return err
	}

	core.Log(next.String())
	return fmt.Errorf("implement releaseStart")
}

// Run the release finish command for the standard workflow.
func (p *standardPlugIn) releaseFinish(_ core.Repository) error {
	return fmt.Errorf("implement releaseFinish")
}

// Run the hotfix start command for the standard workflow.
func (p *standardPlugIn) hotfixStart(repo core.Repository) error {
	// read out the current and next project version ${major}.${minor}.${increment}-${qualifier}
	_, next, err := p.Version(repo.Local(), false, false, true)

	if err != nil {
		return err
	}

	core.Log(next.String())
	return fmt.Errorf("implement hotfixStart")
}

// Run the hotfix finish command for the standard workflow.
func (p *standardPlugIn) hotfixFinish(_ core.Repository) error {
	return fmt.Errorf("implement hotfixFinish")
}
