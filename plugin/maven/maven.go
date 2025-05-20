/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package maven

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mercedes-benz/gitflow-cli/plugin/core"
)

// NewPlugin create plugin for the mvn build tool.
func NewPlugin() core.Plugin {
	plugin := &mavenPlugin{
		majorVersion:           []string{helper, evaluate, fmt.Sprintf(expression, major), quiet, stdout},
		minorVersion:           []string{helper, evaluate, fmt.Sprintf(expression, minor), quiet, stdout},
		incrementalVersion:     []string{helper, evaluate, fmt.Sprintf(expression, incremental), quiet, stdout},
		qualifier:              []string{helper, evaluate, fmt.Sprintf(expression, qualifier), quiet, stdout},
		nextMajorVersion:       []string{helper, evaluate, fmt.Sprintf(expression, nextMajor), quiet, stdout},
		nextMinorVersion:       []string{helper, evaluate, fmt.Sprintf(expression, nextMinor), quiet, stdout},
		nextIncrementalVersion: []string{helper, evaluate, fmt.Sprintf(expression, nextIncremental), quiet, stdout},
		setVersion:             []string{versions, noBackups},
		useReleases:            []string{releases, noBackups, failNotReplaced},
	}

	// RegisterPlugin hooks dynamically for this plugin
	core.GlobalHooks.RegisterHook(pluginName, core.ReleaseStartHooks.AfterUpdateProjectVersionHook, plugin.afterUpdateProjectVersion)

	return plugin
}

// RegisterPlugin plugin for the mvn build tool.
func init() {
	core.RegisterPlugin(NewPlugin())
}

// Name of the mvn plugin.
const pluginName = "Maven"

// Precondition file pluginName for mvn projects.
const preconditionFile = "pom.xml"

// Snapshot qualifier for mvn projects.
const snapshotQualifier = "SNAPSHOT"

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

func (p *mavenPlugin) SnapshotQualifier() string {
	return snapshotQualifier
}

// Maven build tool commands.
const (
	helper          = "build-helper:parse-version"
	evaluate        = "help:evaluate"
	versions        = "versions:set"
	releases        = "versions:use-releases"
	newVersion      = "-DnewVersion=%v"
	expression      = "-Dexpression=parsedVersion.%v"
	major           = "majorVersion"
	minor           = "minorVersion"
	incremental     = "incrementalVersion"
	qualifier       = "qualifier"
	nextMajor       = "nextMajorVersion"
	nextMinor       = "nextMinorVersion"
	nextIncremental = "nextIncrementalVersion"
	quiet           = "-q"
	stdout          = "-DforceStdout"
	noBackups       = "-DgenerateBackupPoms=false"
	failNotReplaced = "-DfailIfNotReplaced=true"
)

// MavenPlugIn is the plugin for the mvn build tool.
type mavenPlugin struct {
	majorVersion           []string
	minorVersion           []string
	incrementalVersion     []string
	qualifier              []string
	nextMajorVersion       []string
	nextMinorVersion       []string
	nextIncrementalVersion []string
	setVersion             []string
	useReleases            []string
}

// CheckRequiredFile Check if the plugin can be executed in a project directory.
func (p *mavenPlugin) CheckRequiredFile(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, preconditionFile))
	return !os.IsNotExist(err)
}

// Version the current and next version of the mvn project.
func (p *mavenPlugin) Version(projectPath string, major, minor, incremental bool) (core.Version, core.Version, error) {
	var currentMajor, currentMinor, currentIncremental, qualifier, nextMajor, nextMinor, nextIncremental string
	var logs []any = make([]any, 0)

	// log human-readable description of the git command
	defer func() { core.Log(logs...) }()

	// evaluate the major version of the mvn project
	majorCommand := exec.Command(Maven, p.majorVersion...)
	majorCommand.Dir = projectPath

	// evaluate the minor version of the mvn project
	minorCommand := exec.Command(Maven, p.minorVersion...)
	minorCommand.Dir = projectPath

	// evaluate the incremental version of the mvn project
	incrementalCommand := exec.Command(Maven, p.incrementalVersion...)
	incrementalCommand.Dir = projectPath

	// evaluate the qualifier of the mvn project
	qualifierCommand := exec.Command(Maven, p.qualifier...)
	qualifierCommand.Dir = projectPath

	// evaluate the next major version of the mvn project
	nextMajorCommand := exec.Command(Maven, p.nextMajorVersion...)
	nextMajorCommand.Dir = projectPath

	// evaluate the next minor version of the mvn project
	nextMinorCommand := exec.Command(Maven, p.nextMinorVersion...)
	nextMinorCommand.Dir = projectPath

	// evaluate the next incremental version of the mvn project
	nextIncrementalCommand := exec.Command(Maven, p.nextIncrementalVersion...)
	nextIncrementalCommand.Dir = projectPath

	// run mvn to evaluate the major version of the mvn project
	if output, err := majorCommand.CombinedOutput(); err != nil {
		logs = append(logs, majorCommand, output, err)

		return core.NoVersion, core.NoVersion,
			fmt.Errorf("mvn major version evaluation failed with %v: %s", err, output)
	} else {
		logs = append(logs, majorCommand, output)
		currentMajor = string(output)
	}

	// run mvn to evaluate the minor version of the mvn project
	if output, err := minorCommand.CombinedOutput(); err != nil {
		logs = append(logs, minorCommand, output, err)

		return core.NoVersion, core.NoVersion,
			fmt.Errorf("mvn minor version evaluation failed with %v: %s", err, output)
	} else {
		logs = append(logs, minorCommand, output)
		currentMinor = string(output)
	}

	// run mvn to evaluate the incremental version of the mvn project
	if output, err := incrementalCommand.CombinedOutput(); err != nil {
		logs = append(logs, incrementalCommand, output, err)

		return core.NoVersion, core.NoVersion,
			fmt.Errorf("mvn incremental version evaluation failed with %v: %s", err, output)
	} else {
		logs = append(logs, incrementalCommand, output)
		currentIncremental = string(output)
	}

	// run mvn to evaluate the next major version of the mvn project
	if output, err := nextMajorCommand.CombinedOutput(); err != nil {
		logs = append(logs, nextMajorCommand, output, err)

		return core.NoVersion, core.NoVersion,
			fmt.Errorf("mvn next major version evaluation failed with %v: %s", err, output)
	} else {
		logs = append(logs, nextMajorCommand, output)
		nextMajor = string(output)
	}

	// run mvn to evaluate the next minor version of the mvn project
	if output, err := nextMinorCommand.CombinedOutput(); err != nil {
		logs = append(logs, nextMinorCommand, output, err)

		return core.NoVersion, core.NoVersion,
			fmt.Errorf("mvn next minor version evaluation failed with %v: %s", err, output)
	} else {
		logs = append(logs, nextMinorCommand, output)
		nextMinor = string(output)
	}

	// run mvn to evaluate the next incremental version of the mvn project
	if output, err := nextIncrementalCommand.CombinedOutput(); err != nil {
		logs = append(logs, nextIncrementalCommand, output, err)

		return core.NoVersion, core.NoVersion,
			fmt.Errorf("mvn next incremental version evaluation failed with %v: %s", err, output)
	} else {
		logs = append(logs, nextIncrementalCommand, output)
		nextIncremental = string(output)
	}

	// run mvn to evaluate the qualifier of the mvn project
	if output, err := qualifierCommand.CombinedOutput(); err != nil {
		logs = append(logs, qualifierCommand, output, err)

		return core.NoVersion, core.NoVersion,
			fmt.Errorf("mvn qualifier evaluation failed with %v: %s", err, output)
	} else {
		logs = append(logs, qualifierCommand, output)
		qualifier = string(output)
	}

	// current and next version of the mvn project
	var nextVersion core.Version
	currentVersion := core.NewVersion(currentMajor, currentMinor, currentIncremental, qualifier)

	// create the next version of the mvn project based on the version increment type
	switch {
	case major && !minor && !incremental:
		// create the next major version of the mvn project
		nextVersion, _ = currentVersion.Increment(core.Major, nextMajor, nextMinor, nextIncremental)

	case minor && !major && !incremental:
		// create the next minor version of the mvn project
		nextVersion, _ = currentVersion.Increment(core.Minor, nextMajor, nextMinor, nextIncremental)

	case incremental && !major && !minor:
		// create the next incremental version of the mvn project
		nextVersion, _ = currentVersion.Increment(core.Incremental, nextMajor, nextMinor, nextIncremental)

	case !major && !minor && !incremental:
		// version increment type not specified, return the current version as next version
		nextVersion = currentVersion

	default:
		return core.NoVersion, core.NoVersion, fmt.Errorf("unsupported version increment type")
	}

	return currentVersion, nextVersion, nil
}

func (p *mavenPlugin) UpdateProjectVersion(next core.Version) error {
	var err error
	var versionCommand *exec.Cmd
	var output []byte

	// log human-readable description of the mvn command
	defer func() { core.Log(versionCommand, output, err) }()

	// update version information
	versionCommand = exec.Command(Maven, append(p.setVersion, fmt.Sprintf(newVersion, next))...)

	// run mvn to update version information of the mvn project
	if output, err = versionCommand.CombinedOutput(); err != nil {
		return fmt.Errorf("mvn versions update failed with %v: %s", err, output)
	}

	return nil
}

// afterUpdateProjectVersion is executed after updating the project version
func (p *mavenPlugin) afterUpdateProjectVersion(plugin core.Plugin, repository core.Repository) error {
	fmt.Println("AfterHook Update Project Version Hook")

	var err error
	var releasesCommand *exec.Cmd
	var output []byte

	// log human-readable description of the mvn command
	defer func() { core.Log(releasesCommand, output, err) }()
	// replace -SNAPSHOT versions and fail if not replaced (i.e. if the version has not been released)
	releasesCommand = exec.Command(Maven, p.useReleases...)
	// todo: implement repository.ProjektPath
	//releasesCommand.Dir = projectPath

	// run mvn to replace -SNAPSHOT versions with releases in the mvn project
	if output, err = releasesCommand.CombinedOutput(); err != nil {
		return fmt.Errorf("mvn releases update failed with %v: %s", err, output)
	}
	return nil
}
