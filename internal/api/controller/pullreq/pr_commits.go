// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pullreq

import (
	"context"
	"fmt"

	"github.com/harness/gitness/gitrpc"
	"github.com/harness/gitness/internal/api/controller"
	"github.com/harness/gitness/internal/auth"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"
)

// Commits lists all commits from pr head branch.
func (c *Controller) Commits(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	pullreqNum int64,
	filter *types.PaginationFilter,
) ([]types.Commit, error) {
	repo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoView)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire access to repo: %w", err)
	}

	pr, err := c.pullreqStore.FindByNumber(ctx, repo.ID, pullreqNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request by number: %w", err)
	}

	gitRef := pr.SourceSHA
	afterRef := pr.MergeBaseSHA

	rpcOut, err := c.gitRPCClient.ListCommits(ctx, &gitrpc.ListCommitsParams{
		ReadParams: gitrpc.CreateRPCReadParams(repo),
		GitREF:     gitRef,
		After:      afterRef,
		Page:       int32(filter.Page),
		Limit:      int32(filter.Limit),
	})
	if err != nil {
		return nil, err
	}

	commits := make([]types.Commit, len(rpcOut.Commits))
	for i := range rpcOut.Commits {
		var commit *types.Commit
		commit, err = controller.MapCommit(&rpcOut.Commits[i])
		if err != nil {
			return nil, fmt.Errorf("failed to map commit: %w", err)
		}
		commits[i] = *commit
	}

	return commits, nil
}
