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
	// check if the repository already has a release branch
	if found, _, err := repo.HasBranch(core.Release); err != nil {
		return err
	} else if found {
		return fmt.Errorf(
			"repository already has a '%v' branch and only one '%v' branch is allowed at a time",
			core.Release, core.Release)
	}

	// check if the repository has a develop branch // todo: has remote branch?
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
	//   set the version of project to (${major}+1).0.0-${qualifier}
	//   perform a git commit with a commit message
	if next.VersionIncrement == core.Major {
		if err := p.updateProjectVersion(next.AddQualifier(snapshotQualifier)); err != nil {
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
	if err := p.updateProjectVersion(current.RemoveQualifier()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repo.CommitChanges("Remove qualifier from project version."); err != nil {
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

// Run the release finish command for the standard workflow.
func (p *standardPlugIn) releaseFinish(repo core.Repository) error {
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
	} else if err := p.updateProjectVersion(next.AddQualifier(snapshotQualifier)); err != nil {
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

// Run the hotfix start command for the standard workflow.
func (p *standardPlugIn) hotfixStart(repo core.Repository) error {
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
	if err := p.updateProjectVersion(next); err != nil {
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

// Run the hotfix finish command for the standard workflow.
func (p *standardPlugIn) hotfixFinish(repo core.Repository) error {
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
		if err := p.updateProjectVersion(hotfixVersion); err != nil {
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
		// change version im develop Branch to the previous snapshot version (or find a better solution)
		// but at the end the project version in develop branch should remain the same as before hotfix merge
		if err := p.updateProjectVersion(currentVersion); err != nil {
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

// Sets the project's version
func (p *standardPlugIn) updateProjectVersion(next core.Version) error {

	if err := os.WriteFile(preconditionFile, []byte(next.String()), 0644); err != nil {
		return fmt.Errorf("failed to write in file %v next project version %v", preconditionFile, next.String())
	}

	return nil
}
