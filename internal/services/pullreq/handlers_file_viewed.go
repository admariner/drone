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
	"errors"
	"fmt"
	"io"

	"github.com/harness/gitness/events"
	"github.com/harness/gitness/gitrpc"
	pullreqevents "github.com/harness/gitness/internal/events/pullreq"
)

// handleFileViewedOnBranchUpdate handles pull request Branch Updated events.
// It marks existing file reviews as obsolete for the PR depending on the change to the file.
//
// The major reason of this handler is to allow detect changes that occured to a file since last reviewed,
// even if the file content is the same - e.g. file got deleted and readded with the same content.
func (s *Service) handleFileViewedOnBranchUpdate(ctx context.Context,
	event *events.Event[*pullreqevents.BranchUpdatedPayload],
) error {
	repoGit, err := s.repoGitInfoCache.Get(ctx, event.Payload.TargetRepoID)
	if err != nil {
		return fmt.Errorf("failed to get repo git info: %w", err)
	}
	reader := gitrpc.NewStreamReader(s.gitRPCClient.Diff(ctx, &gitrpc.DiffParams{
		ReadParams: gitrpc.ReadParams{
			RepoUID: repoGit.GitUID,
		},
		BaseRef:      event.Payload.OldSHA,
		HeadRef:      event.Payload.NewSHA,
		MergeBase:    false, // we want the direct changes
		IncludePatch: false, // we don't care about the actual file changes
	}))

	obsoletePaths := []string{}
	for {
		fileDiff, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read next file diff: %w", err)
		}

		// DELETED: mark as obsolete - handles open pr file deletions
		// CREATED: mark as obsolete - handles cases in which file deleted while PR was closed
		// RENAMED: mark old + new path as obsolete - similar to deleting old file and creating new one
		// UPDATED: mark as obsolete - in case pr is closed file SHA is handling it
		// This strategy leads to a behavior very similar to what github is doing
		switch fileDiff.Status {
		case gitrpc.FileDiffStatusAdded:
			obsoletePaths = append(obsoletePaths, fileDiff.Path)
		case gitrpc.FileDiffStatusDeleted:
			obsoletePaths = append(obsoletePaths, fileDiff.OldPath)
		case gitrpc.FileDiffStatusRenamed:
			obsoletePaths = append(obsoletePaths, fileDiff.OldPath, fileDiff.Path)
		case gitrpc.FileDiffStatusModified:
			obsoletePaths = append(obsoletePaths, fileDiff.Path)
		default:
			// other cases we don't care
		}
	}

	if len(obsoletePaths) == 0 {
		return nil
	}

	err = s.fileViewStore.MarkObsolete(
		ctx,
		event.Payload.PullReqID,
		obsoletePaths)
	if err != nil {
		return fmt.Errorf(
			"failed to mark files obsolete for repo %d and pr %d: %w",
			repoGit.ID,
			event.Payload.PullReqID,
			err)
	}

	return nil
}
