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

package router

import (
	"github.com/harness/gitness/gitrpc"
	"github.com/harness/gitness/internal/api/controller/check"
	"github.com/harness/gitness/internal/api/controller/connector"
	"github.com/harness/gitness/internal/api/controller/execution"
	"github.com/harness/gitness/internal/api/controller/githook"
	"github.com/harness/gitness/internal/api/controller/logs"
	"github.com/harness/gitness/internal/api/controller/pipeline"
	"github.com/harness/gitness/internal/api/controller/plugin"
	"github.com/harness/gitness/internal/api/controller/principal"
	"github.com/harness/gitness/internal/api/controller/pullreq"
	"github.com/harness/gitness/internal/api/controller/repo"
	"github.com/harness/gitness/internal/api/controller/secret"
	"github.com/harness/gitness/internal/api/controller/serviceaccount"
	"github.com/harness/gitness/internal/api/controller/space"
	"github.com/harness/gitness/internal/api/controller/system"
	"github.com/harness/gitness/internal/api/controller/template"
	"github.com/harness/gitness/internal/api/controller/trigger"
	"github.com/harness/gitness/internal/api/controller/user"
	"github.com/harness/gitness/internal/api/controller/webhook"
	"github.com/harness/gitness/internal/auth/authn"
	"github.com/harness/gitness/internal/auth/authz"
	"github.com/harness/gitness/internal/store"
	"github.com/harness/gitness/internal/url"
	"github.com/harness/gitness/types"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideRouter,
	ProvideGitHandler,
	ProvideAPIHandler,
	ProvideWebHandler,
)

func ProvideRouter(
	config *types.Config,
	api APIHandler,
	git GitHandler,
	web WebHandler,
) *Router {
	return NewRouter(api, git, web,
		config.Server.HTTP.GitHost)
}

func ProvideGitHandler(
	config *types.Config,
	urlProvider *url.Provider,
	repoStore store.RepoStore,
	authenticator authn.Authenticator,
	authorizer authz.Authorizer,
	client gitrpc.Interface,
) GitHandler {
	return NewGitHandler(config, urlProvider, repoStore, authenticator, authorizer, client)
}

func ProvideAPIHandler(
	config *types.Config,
	authenticator authn.Authenticator,
	repoCtrl *repo.Controller,
	executionCtrl *execution.Controller,
	logCtrl *logs.Controller,
	spaceCtrl *space.Controller,
	pipelineCtrl *pipeline.Controller,
	secretCtrl *secret.Controller,
	triggerCtrl *trigger.Controller,
	connectorCtrl *connector.Controller,
	templateCtrl *template.Controller,
	pluginCtrl *plugin.Controller,
	pullreqCtrl *pullreq.Controller,
	webhookCtrl *webhook.Controller,
	githookCtrl *githook.Controller,
	saCtrl *serviceaccount.Controller,
	userCtrl *user.Controller,
	principalCtrl principal.Controller,
	checkCtrl *check.Controller,
	sysCtrl *system.Controller,
) APIHandler {
	return NewAPIHandler(config, authenticator, repoCtrl, executionCtrl, logCtrl, spaceCtrl, pipelineCtrl,
		secretCtrl, triggerCtrl, connectorCtrl, templateCtrl, pluginCtrl, pullreqCtrl, webhookCtrl,
		githookCtrl, saCtrl, userCtrl, principalCtrl, checkCtrl, sysCtrl)
}

func ProvideWebHandler(config *types.Config) WebHandler {
	return NewWebHandler(config)
}
