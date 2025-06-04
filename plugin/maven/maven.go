/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package maven

import (
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/plugin/core"
	"os/exec"
	"strings"
)

// NewPlugin create plugin for the mvn build tool.
func NewPlugin() core.Plugin {
	plugin := &mavenPlugin{
		getVersion:  []string{evaluate, versionProperty, quiet, stdout},
		setVersion:  []string{versions, noBackups},
		useReleases: []string{releases, noBackups, failNotReplaced},
	}

	// RegisterPlugin hooks dynamically for this plugin
	core.GlobalHooks.RegisterHook(pluginName, core.ReleaseStartHooks.AfterUpdateProjectVersionHook, plugin.afterUpdateProjectVersion)

	return plugin
}

// RegisterPlugin plugin for the mvn build tool.
func init() {
	core.RegisterPlugin(NewPlugin())
}

const pluginName = "Maven"

const versionFileName = "pom.xml"

const versionQualifier = "SNAPSHOT"

const (
	Maven = "mvn"
)

// RequiredTools list of required command line tools
func (p *mavenPlugin) RequiredTools() []string {
	return []string{Maven}
}

func (p *mavenPlugin) String() string {
	return pluginName
}

func (p *mavenPlugin) VersionFileName() string {
	return versionFileName
}

func (p *mavenPlugin) VersionQualifier() string {
	return versionQualifier
}

// Maven build tool commands.
const (
	evaluate        = "help:evaluate"
	versions        = "versions:set"
	releases        = "versions:use-releases"
	newVersion      = "-DnewVersion=%v"
	versionProperty = "-Dexpression=project.version"
	quiet           = "-q"
	stdout          = "-DforceStdout"
	noBackups       = "-DgenerateBackupPoms=false"
	failNotReplaced = "-DfailIfNotReplaced=true"
)

// MavenPlugIn is the plugin for the mvn build tool.
type mavenPlugin struct {
	getVersion  []string
	setVersion  []string
	useReleases []string
}

// ReadVersion reads the current version from the project
func (p *mavenPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	var logs []any = make([]any, 0)
	projectPath := repository.Local()

	// log human-readable description of the git command
	defer func() { core.Log(logs...) }()

	// evaluate the version of the mvn project
	versionCommand := exec.Command(Maven, p.getVersion...)
	versionCommand.Dir = projectPath

	// run mvn to evaluate the version of the mvn project
	output, err := versionCommand.CombinedOutput()
	if err != nil {
		logs = append(logs, versionCommand, output, err)
		return core.NoVersion, fmt.Errorf("mvn version evaluation failed with %v: %s", err, output)
	}

	logs = append(logs, versionCommand, output)
	versionStr := strings.TrimSpace(string(output))

	// parse the version string using core.ParseVersion
	return core.ParseVersion(versionStr)
}

// WriteVersion writes a new version to the project
func (p *mavenPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	var err error
	var versionCommand *exec.Cmd
	var output []byte
	projectPath := repository.Local()

	// log human-readable description of the mvn command
	defer func() { core.Log(versionCommand, output, err) }()

	// update version information
	versionCommand = exec.Command(Maven, append(p.setVersion, fmt.Sprintf(newVersion, version))...)
	versionCommand.Dir = projectPath

	// run mvn to update version information of the mvn project
	if output, err = versionCommand.CombinedOutput(); err != nil {
		return fmt.Errorf("mvn versions update failed with %v: %s", err, output)
	}

	return nil
}

// afterUpdateProjectVersion is executed after updating the project version
func (p *mavenPlugin) afterUpdateProjectVersion(repository core.Repository) error {
	fmt.Println("After Update Project Version Hook")

	var err error
	var releasesCommand *exec.Cmd
	var output []byte

	// log human-readable description of the mvn command
	defer func() { core.Log(releasesCommand, output, err) }()
	// replace -SNAPSHOT versions and fail if not replaced (i.e. if the version has not been released)
	releasesCommand = exec.Command(Maven, p.useReleases...)
	releasesCommand.Dir = repository.Local()

	// run mvn to replace -SNAPSHOT versions with releases in the mvn project
	if output, err = releasesCommand.CombinedOutput(); err != nil {
		return fmt.Errorf("mvn releases update failed with %v: %s", err, output)
	}

	// if not clean: perform a git commit with a commit message because the previous step changed the POM file
	if err := repository.IsClean(); err != nil {
		if err := repository.CommitChanges("Update project dependencies with corresponding releases."); err != nil {
			return repository.UndoAllChanges(err)
		}
	}
	return nil
}
