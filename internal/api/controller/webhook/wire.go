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

package webhook

import (
	"github.com/harness/gitness/encrypt"
	"github.com/harness/gitness/internal/auth/authz"
	"github.com/harness/gitness/internal/services/webhook"
	"github.com/harness/gitness/internal/store"

	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(config webhook.Config, db *sqlx.DB, authorizer authz.Authorizer,
	webhookStore store.WebhookStore, webhookExecutionStore store.WebhookExecutionStore,
	repoStore store.RepoStore, webhookService *webhook.Service, encrypter encrypt.Encrypter) *Controller {
	return NewController(config.AllowLoopback, config.AllowPrivateNetwork,
		db, authorizer, webhookStore, webhookExecutionStore, repoStore, webhookService, encrypter)
}
