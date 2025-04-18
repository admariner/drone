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

package githook

import (
	"context"
	"fmt"

	"github.com/harness/gitness/internal/api/usererror"
	"github.com/harness/gitness/internal/auth"
	"github.com/harness/gitness/internal/auth/authz"
	eventsgit "github.com/harness/gitness/internal/events/git"
	"github.com/harness/gitness/internal/store"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"github.com/jmoiron/sqlx"
)

// ServerHookOutput represents the output of server hook api calls.
// TODO: support non-error messages (once we need it).
type ServerHookOutput struct {
	// Error contains the user facing error (like "branch is protected", ...).
	Error *string `json:"error,omitempty"`
}

// ReferenceUpdate represents an update of a git reference.
type ReferenceUpdate struct {
	// Ref is the full name of the reference that got updated.
	Ref string `json:"ref"`
	// Old is the old commmit hash (before the update).
	Old string `json:"old"`
	// New is the new commit hash (after the update).
	New string `json:"new"`
}

// BaseInput contains the base input for any githook api call.
type BaseInput struct {
	RepoID      int64 `json:"repo_id"`
	PrincipalID int64 `json:"principal_id"`
}

type Controller struct {
	db             *sqlx.DB
	authorizer     authz.Authorizer
	principalStore store.PrincipalStore
	repoStore      store.RepoStore
	gitReporter    *eventsgit.Reporter
}

func NewController(
	db *sqlx.DB,
	authorizer authz.Authorizer,
	principalStore store.PrincipalStore,
	repoStore store.RepoStore,
	gitReporter *eventsgit.Reporter,
) *Controller {
	return &Controller{
		db:             db,
		authorizer:     authorizer,
		principalStore: principalStore,
		repoStore:      repoStore,
		gitReporter:    gitReporter,
	}
}

func (c *Controller) getRepoCheckAccess(ctx context.Context,
	_ *auth.Session, repoID int64, _ enum.Permission) (*types.Repository, error) {
	if repoID < 1 {
		return nil, usererror.BadRequest("A valid repository reference must be provided.")
	}

	repo, err := c.repoStore.Find(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to find repo with id %d: %w", repoID, err)
	}

	// TODO: execute permission check. block anything but gitness service?

	return repo, nil
}
