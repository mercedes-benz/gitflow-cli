/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import "fmt"

// BranchSyncFunc is called when a configured branch name doesn't match any remote branch.
// It receives context about the situation and returns the resolved branch name (empty = abort).
type BranchSyncFunc func(request BranchSyncRequest) (BranchSyncResult, error)

// BranchSyncRequest describes the mismatch between configured and actual branches.
type BranchSyncRequest struct {
	BranchType Branch
	Configured string
	Candidates []string
	CreateFrom string
	Repository Repository
}

// BranchSyncResult describes how to resolve the branch mismatch.
type BranchSyncResult struct {
	ResolvedName string
	Created      bool
	Persist      bool
}

// BranchSync is the global callback for resolving branch mismatches.
// If nil, an error is returned when branches don't match.
var BranchSync BranchSyncFunc

var branchCandidates = map[Branch][]string{
	Production:  {"main", "master"},
	Development: {"develop", "dev", "development"},
}

// syncBranch checks that the configured branch exists on remote.
// If not found, it invokes BranchSync to offer resolution.
func syncBranch(repository Repository, branchType Branch) error {
	found, _, err := repository.HasBranch(branchType)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	candidates := findCandidates(repository, branchType)

	if BranchSync == nil {
		if len(candidates) > 0 {
			return fmt.Errorf("branch '%v' not found (did you mean '%s'?)", branchType, candidates[0])
		}
		return fmt.Errorf("repository does not have a '%v' branch", branchType)
	}

	createFrom := ""
	if branchType == Development {
		createFrom = branchNames[Production]
	}

	result, err := BranchSync(BranchSyncRequest{
		BranchType: branchType,
		Configured: branchNames[branchType],
		Candidates: candidates,
		CreateFrom: createFrom,
		Repository: repository,
	})
	if err != nil {
		return err
	}
	if result.ResolvedName == "" {
		return fmt.Errorf("branch '%v' is required but was not resolved", branchType)
	}

	if result.Created {
		if err := repository.CheckoutBranch(createFrom); err != nil {
			return err
		}
		if err := repository.CreateBranch(result.ResolvedName); err != nil {
			return err
		}
		if err := pushIfEnabled(func() error {
			return repository.PushChanges(result.ResolvedName)
		}); err != nil {
			return err
		}
	}

	branchNames[branchType] = result.ResolvedName
	return nil
}

func findCandidates(repository Repository, branchType Branch) []string {
	configured := branchNames[branchType]
	var found []string
	for _, candidate := range branchCandidates[branchType] {
		if candidate == configured {
			continue
		}
		if exists, err := repository.HasRemoteBranch(candidate); err == nil && exists {
			found = append(found, candidate)
		}
	}
	return found
}
