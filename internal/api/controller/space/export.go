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

package space

import (
	"context"
	"errors"
	"fmt"

	apiauth "github.com/harness/gitness/internal/api/auth"
	"github.com/harness/gitness/internal/api/usererror"
	"github.com/harness/gitness/internal/auth"
	"github.com/harness/gitness/internal/services/exporter"
	"github.com/harness/gitness/store/database/dbtx"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"
)

type ExportInput struct {
	AccountId         string `json:"accountId"`
	OrgIdentifier     string `json:"orgIdentifier"`
	ProjectIdentifier string `json:"projectIdentifier"`
	Token             string `json:"token"`
}

// Export creates a new empty repository in harness code and does git push to it.
func (c *Controller) Export(ctx context.Context, session *auth.Session, spaceRef string, in *ExportInput) error {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return err
	}

	if err = apiauth.CheckSpace(ctx, c.authorizer, session, space, enum.PermissionSpaceEdit, false); err != nil {
		return err
	}

	err = c.sanitizeExportInput(in)
	if err != nil {
		return fmt.Errorf("failed to sanitize input: %w", err)
	}

	providerInfo := &exporter.HarnessCodeInfo{
		AccountId:         in.AccountId,
		ProjectIdentifier: in.ProjectIdentifier,
		OrgIdentifier:     in.OrgIdentifier,
		Token:             in.Token,
	}

	var repos []*types.Repository
	page := 1
	for {
		reposInPage, err := c.repoStore.List(ctx, space.ID, &types.RepoFilter{Size: 200, Page: page, Order: enum.OrderDesc})
		if err != nil {
			return err
		}
		if len(reposInPage) == 0 {
			break
		}
		page += 1
		repos = append(repos, reposInPage...)
	}

	err = dbtx.New(c.db).WithTx(ctx, func(ctx context.Context) error {
		err = c.exporter.RunManyForSpace(ctx, space.ID, repos, providerInfo)
		if errors.Is(err, exporter.ErrJobRunning) {
			return usererror.ConflictWithPayload("export already in progress")
		}
		if err != nil {
			return fmt.Errorf("failed to start export repository job: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) sanitizeExportInput(in *ExportInput) error {
	if in.AccountId == "" {
		return usererror.BadRequest("account id must be provided")
	}

	if in.OrgIdentifier == "" {
		return usererror.BadRequest("organization identifier must be provided")
	}

	if in.ProjectIdentifier == "" {
		return usererror.BadRequest("project identifier must be provided")
	}

	if in.Token == "" {
		return usererror.BadRequest("token for harness code must be provided")
	}

	return nil
}
