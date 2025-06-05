/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package npm

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mercedes-benz/gitflow-cli/plugin/core"
)

// Required tools for the npm plugin.
const (
	npm = "npm"
)

// Versioning file for package.json projects.
const (
	versionFile      = "package.json"
	versionQualifier = "dev"
)

// npmPlugin is the plugin implementation for npm projects.
var npmPlugin = &plugin{}

// init registers the npm plugin.
func init() {
	core.RegisterPlugin(npmPlugin)
}

// plugin is the struct implementing the Plugin interface.
type plugin struct{}

// String returns the name of the plugin.
func (p *plugin) String() string {
	return "NPM"
}

// VersionFileName returns the filename containing version information.
func (p *plugin) VersionFileName() string {
	return versionFile
}

// VersionQualifier returns the qualifier for version strings.
func (p *plugin) VersionQualifier() string {
	return versionQualifier
}

// RequiredTools returns the list of required tools for this plugin.
func (p *plugin) RequiredTools() []string {
	return []string{npm}
}

// ReadVersion reads the version from package.json using npm.
func (p *plugin) ReadVersion(repository core.Repository) (core.Version, error) {
	// Execute npm command to read the version from package.json
	cmd := exec.Command(npm, "pkg", "get", "version")
	cmd.Dir = repository.Local()

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return core.Version{}, fmt.Errorf("failed to read version: %v", err)
	}

	// Clean the version string
	versionString := strings.TrimSpace(stdout.String())
	// Remove surrounding quotes from the npm output
	versionString = strings.Trim(versionString, "\"")

	// Parse the version string
	version, err := core.ParseVersion(versionString)
	if err != nil {
		return core.Version{}, fmt.Errorf("failed to parse version: %v", err)
	}

	return version, nil
}

// WriteVersion writes the version to package.json using npm.
func (p *plugin) WriteVersion(repository core.Repository, version core.Version) error {
	// Execute npm command to write the version to package.json
	cmd := exec.Command(npm, "version", version.String(), "--no-git-tag-version")
	cmd.Dir = repository.Local()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write version: %v", err)
	}

	return nil
}
