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

// NewPlugIn create plugin for the mvn build tool.
func NewPlugIn() core.PlugIn {
	return &mavenPlugIn{
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
}

// Name of the mvn plugin.
const name = "Maven"

// Precondition file name for mvn projects.
const preconditionFile = "pom.xml"

// Snapshot qualifier for mvn projects.
const snapshotQualifier = "SNAPSHOT"

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
type mavenPlugIn struct {
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

// Check if the plugin can be executed in a project directory.
func (p *mavenPlugIn) Check(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, preconditionFile))
	return !os.IsNotExist(err)
}

// Version the current and next version of the mvn project.
func (p *mavenPlugIn) Version(projectPath string, major, minor, incremental bool) (core.Version, core.Version, error) {
	var currentMajor, currentMinor, currentIncremental, qualifier, nextMajor, nextMinor, nextIncremental string
	var logs []any = make([]any, 0)

	// log human-readable description of the git command
	defer func() { core.Log(logs...) }()

	// evaluate the major version of the mvn project
	majorCommand := exec.Command(core.Maven, p.majorVersion...)
	majorCommand.Dir = projectPath

	// evaluate the minor version of the mvn project
	minorCommand := exec.Command(core.Maven, p.minorVersion...)
	minorCommand.Dir = projectPath

	// evaluate the incremental version of the mvn project
	incrementalCommand := exec.Command(core.Maven, p.incrementalVersion...)
	incrementalCommand.Dir = projectPath

	// evaluate the qualifier of the mvn project
	qualifierCommand := exec.Command(core.Maven, p.qualifier...)
	qualifierCommand.Dir = projectPath

	// evaluate the next major version of the mvn project
	nextMajorCommand := exec.Command(core.Maven, p.nextMajorVersion...)
	nextMajorCommand.Dir = projectPath

	// evaluate the next minor version of the mvn project
	nextMinorCommand := exec.Command(core.Maven, p.nextMinorVersion...)
	nextMinorCommand.Dir = projectPath

	// evaluate the next incremental version of the mvn project
	nextIncrementalCommand := exec.Command(core.Maven, p.nextIncrementalVersion...)
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

// Start command of the mvn build tool.
func (p *mavenPlugIn) Start(branch core.Branch, projectPath string, args ...any) error {
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
	return core.StartWorkflow(start, []string{core.Git, core.Maven}, branch, name, projectPath, args...)
}

// Finish command of the mvn build tool.
func (p *mavenPlugIn) Finish(branch core.Branch, projectPath string) error {
	// select suitable business logic for the branch
	switch branch {
	case core.Release:
		// finish the workflow with the selected release business logic
		return core.FinishWorkflow(p.releaseFinish, []string{core.Git, core.Maven}, branch, name, projectPath)

	case core.Hotfix:
		// finish the workflow with the selected hotfix business logic
		return core.FinishWorkflow(p.hotfixFinish, []string{core.Git, core.Maven}, branch, name, projectPath)

	default:
		return fmt.Errorf("unsupported branch: %v", branch)
	}
}

// Register plugin for the mvn build tool.
func init() {
	core.Register(NewPlugIn())
}

// Run the release start command for the mvn build tool.
func (p *mavenPlugIn) releaseStart(repo core.Repository, major, minor bool) error {
	// check if the repository already has a release branch
	if found, _, err := repo.HasBranch(core.Release); err != nil {
		return err
	} else if found {
		return fmt.Errorf(
			"repository already has a '%v' branch and only one '%v' branch is allowed at a time",
			core.Release, core.Release)
	}

	// check if the repository has a develop branch
	if found, _, err := repo.HasBranch(core.Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to start a new '%v' branch from",
			core.Development, core.Release)
	}

	// checkout develop branch
	if err := repo.CheckoutBranch(core.Development.String()); err != nil {
		return err
	}

	// read out the current and next project version ${major}.${minor}.${increment}-${qualifier}
	current, next, err := p.Version(repo.Local(), major, minor, false)

	if err != nil {
		return err
	}

	// if --major Flag only
	//   set the version of project to (${major}+1).0.0-${qualifier} (change POM file)
	//   perform a git commit with a commit message
	if next.VersionIncrement == core.Major {
		if err := p.updateProjectObjectModelVersion(repo.Local(), next.AddQualifier(snapshotQualifier)); err != nil {
			return repo.UndoAllChanges(err)
		}

		if err := repo.CommitChanges("Set next major project version."); err != nil {
			return repo.UndoAllChanges(err)
		}

		current = next
	}

	// create branch release/x.y.z based on the current develop branch without qualifier
	// checkout release/x.y.z branch
	if err := repo.CreateBranch(current.RemoveQualifier().BranchName(core.Release)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// remove qualifier from the project version (change POM file)
	if err := p.updateProjectObjectModelVersion(repo.Local(), current.RemoveQualifier()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repo.CommitChanges("Remove qualifier from project version."); err != nil {
		return repo.UndoAllChanges(err)
	}

	// execute https://www.mojohaus.org/versions/versions-maven-plugin/use-releases-mojo.html
	if err := p.updateProjectObjectModelReleases(repo.Local()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// if not clean: perform a git commit with a commit message because the previous step changed the POM file
	if err := repo.IsClean(); err != nil {
		if err := repo.CommitChanges("Update project dependencies with corresponding releases."); err != nil {
			return repo.UndoAllChanges(err)
		}
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(core.Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	return nil
}

// Run the release finish command for the mvn build tool.
func (p *mavenPlugIn) releaseFinish(repo core.Repository) error {
	var releaseVersion core.Version

	// check if the repository has a suitable release branch
	if found, remotes, err := repo.HasBranch(core.Release); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("repository does not have a '%v' branch to finish", core.Release)
	} else if len(remotes) > 1 {
		return fmt.Errorf("repository must not have multiple '%v' branches", core.Release)
	} else if version, err := core.ParseVersion(remotes[0]); err != nil {
		return err
	} else {
		releaseVersion = version
	}

	// check if the repository has a develop branch
	if found, _, err := repo.HasBranch(core.Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to finish and merge with a '%v' branch",
			core.Development, core.Release)
	}

	// checkout release branch
	if err := repo.CheckoutBranch(releaseVersion.BranchName(core.Release)); err != nil {
		return err
	}

	// checkout production branch
	if err := repo.CheckoutBranch(core.Production.String()); err != nil {
		return err
	}

	// merge release branch into current production branch (with merge commit --no-ff git flag)
	if err := repo.MergeBranch(releaseVersion.BranchName(core.Release), core.NoFastForward); err != nil {
		return repo.UndoAllChanges(err)
	}

	// tag last commit with the release version number
	if err := repo.TagCommit(releaseVersion.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout develop branch
	if err := repo.CheckoutBranch(core.Development.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// merge release branch into current develop branch (with merge commit --no-ff git flag)
	if err := repo.MergeBranch(releaseVersion.BranchName(core.Release), core.NoFastForward); err != nil {
		return repo.UndoAllChanges(err)
	}

	// set project version to the next develop version ${major}.(${minor}+1).0-${qualifier} (change POM file)
	if _, next, err := p.Version(repo.Local(), false, true, false); err != nil {
		return repo.UndoAllChanges(err)
	} else if err := p.updateProjectObjectModelVersion(repo.Local(), next.AddQualifier(snapshotQualifier)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repo.CommitChanges("Set next minor project version."); err != nil {
		return repo.UndoAllChanges(err)
	}

	// delete the release branch locally
	if err := repo.DeleteBranch(releaseVersion.BranchName(core.Release)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(core.Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	// push all tags to remotes
	if err := repo.PushAllTags(); err != nil {
		return err
	}

	// delete the release branch remotely
	if err := repo.PushDeletion(releaseVersion.BranchName(core.Release)); err != nil {
		return err
	}

	return nil
}

// Run the hotfix start command for the mvn build tool.
func (p *mavenPlugIn) hotfixStart(repo core.Repository) error {
	// check if the repository already has a hotfix branch
	if found, _, err := repo.HasBranch(core.Hotfix); err != nil {
		return err
	} else if found {
		return fmt.Errorf(
			"repository already has a '%v' branch and only one '%v' branch is allowed at a time",
			core.Hotfix, core.Hotfix)
	}

	// checkout production branch
	if err := repo.CheckoutBranch(core.Production.String()); err != nil {
		return err
	}

	// read out the current and next project version ${major}.${minor}.${increment}-${qualifier}
	_, next, err := p.Version(repo.Local(), false, false, true)

	if err != nil {
		return err
	}

	// create branch hotfix/${major}.${minor}.${increment + 1} based on the current production branch
	// checkout hotfix/${major}.${minor}.${increment + 1} branch
	if err := repo.CreateBranch(next.BranchName(core.Hotfix)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// update project version to ${major}.${minor}.${increment + 1}
	if err := p.updateProjectObjectModelVersion(repo.Local(), next); err != nil {
		return repo.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repo.CommitChanges("Set next hotfix version."); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(core.Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	return nil
}

// Run the hotfix finish command for the mvn build tool.
func (p *mavenPlugIn) hotfixFinish(repo core.Repository) error {
	var hotfixVersion core.Version

	// check if the repository has a suitable hotfix branch
	if found, remotes, err := repo.HasBranch(core.Hotfix); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("repository does not have a '%v' branch to finish", core.Hotfix)
	} else if len(remotes) > 1 {
		return fmt.Errorf("repository must not have multiple '%v' branches", core.Hotfix)
	} else if version, err := core.ParseVersion(remotes[0]); err != nil {
		return err
	} else {
		hotfixVersion = version
	}

	// check if the repository has a develop branch
	if found, _, err := repo.HasBranch(core.Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to finish and merge with a '%v' branch",
			core.Development, core.Hotfix)
	}

	// checkout hotfix branch
	if err := repo.CheckoutBranch(hotfixVersion.BranchName(core.Hotfix)); err != nil {
		return err
	}

	// checkout production branch
	if err := repo.CheckoutBranch(core.Production.String()); err != nil {
		return err
	}

	// merge hotfix branch into current production branch (with merge commit --no-ff git flag)
	if err := repo.MergeBranch(hotfixVersion.BranchName(core.Hotfix), core.NoFastForward); err != nil {
		return repo.UndoAllChanges(err)
	}

	// tag last commit with the hotfix version number
	if err := repo.TagCommit(hotfixVersion.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout develop branch
	if err := repo.CheckoutBranch(core.Development.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// in order to avoid merge conflicts, set and commit pom.xml project version in develop branch equal
	// with current hotfix branch and remember its commit hash (or find a better solution)
	if currentVersion, _, err := p.Version(repo.Local(), false, false, false); err != nil {
		return repo.UndoAllChanges(err)
	} else {
		// update project version to ${major}.${minor}.${increment + 1} (means: hotfix branch version)
		if err := p.updateProjectObjectModelVersion(repo.Local(), hotfixVersion); err != nil {
			return repo.UndoAllChanges(err)
		}

		// perform a git commit with a commit message
		if err := repo.CommitChanges("Set hotfix version to avoid merge conflict."); err != nil {
			return repo.UndoAllChanges(err)
		}

		// merge hotfix branch into current develop branch (with merge commit --no-ff git flag)
		if err := repo.MergeBranch(hotfixVersion.BranchName(core.Hotfix), core.NoFastForward); err != nil {
			return repo.UndoAllChanges(err)
		}

		// remove previous commit with remembered commit hash, since it was committed just in order
		// to avoid merge conflicts (or find a better solution)
		// change pom.xml im develop Branch to the previous snapshot version (or find a better solution)
		// but at the end the pom project version in develop branch should remain the same as before hotfix merge
		if err := p.updateProjectObjectModelVersion(repo.Local(), currentVersion); err != nil {
			return repo.UndoAllChanges(err)
		}

		// perform a git commit with a commit message
		if err := repo.CommitChanges("Set version back to project version before hotfix merge."); err != nil {
			return repo.UndoAllChanges(err)
		}
	}

	// delete the release branch locally
	if err := repo.DeleteBranch(hotfixVersion.BranchName(core.Hotfix)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(core.Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	// push all tags to remotes
	if err := repo.PushAllTags(); err != nil {
		return err
	}

	// delete the hotfix branch remotely
	if err := repo.PushDeletion(hotfixVersion.BranchName(core.Hotfix)); err != nil {
		return err
	}

	return nil
}

// Sets the mvn project's version and based on that change propagates that change onto any child modules as necessary.
func (p *mavenPlugIn) updateProjectObjectModelVersion(projectPath string, next core.Version) error {
	var err error
	var versionCommand *exec.Cmd
	var output []byte

	// log human-readable description of the mvn command
	defer func() { core.Log(versionCommand, output, err) }()

	// update version information
	versionCommand = exec.Command(core.Maven, append(p.setVersion, fmt.Sprintf(newVersion, next))...)
	versionCommand.Dir = projectPath

	// run mvn to update version information of the mvn project
	if output, err = versionCommand.CombinedOutput(); err != nil {
		return fmt.Errorf("mvn versions update failed with %v: %s", err, output)
	}

	return nil
}

// Replaces any -SNAPSHOT versions with the corresponding release version (if it has been released).
func (p *mavenPlugIn) updateProjectObjectModelReleases(projectPath string) error {
	var err error
	var releasesCommand *exec.Cmd
	var output []byte

	// log human-readable description of the mvn command
	defer func() { core.Log(releasesCommand, output, err) }()

	// replace -SNAPSHOT versions and fail if not replaced (i.e. if the version has not been released)
	releasesCommand = exec.Command(core.Maven, p.useReleases...)
	releasesCommand.Dir = projectPath

	// run mvn to replace -SNAPSHOT versions with releases in the mvn project
	if output, err = releasesCommand.CombinedOutput(); err != nil {
		return fmt.Errorf("mvn releases update failed with %v: %s", err, output)
	}

	return nil
}
