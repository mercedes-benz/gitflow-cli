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

// NewPlugIn Create plugin for the standard workflow.
func NewPlugIn() core.PlugIn {
	return &standardPlugIn{}
}

// Name of the standard plugin.
const name = "Standard"

// Precondition file name for standard projects.
const preconditionFile = "version.txt"

// Snapshot qualifier for mvn projects.
const snapshotQualifier = "dev"

// StandardPlugIn is the plugin for the standard workflow.
type standardPlugIn struct {
	majorVersion           []string
	minorVersion           []string
	incrementalVersion     []string
	qualifier              []string
	nextMajorVersion       []string
	nextMinorVersion       []string
	nextIncrementalVersion []string
	setVersion             []string
}

func (p *standardPlugIn) Name() string {
	return name
}

func (p *standardPlugIn) SnapshotQualifier() string {
	return snapshotQualifier
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

// Register plugin for the standard workflow.
func init() {
	core.Register(NewPlugIn())
}

// UpdateProjectVersion Sets the project's version
func (p *standardPlugIn) UpdateProjectVersion(next core.Version) error {

	if err := os.WriteFile(preconditionFile, []byte(next.String()), 0644); err != nil {
		return fmt.Errorf("failed to write in file %v next project version %v", preconditionFile, next.String())
	}

	return nil
}
