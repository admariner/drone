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
	"net/http"

	"github.com/harness/gitness/internal/api/openapi"
	"github.com/harness/gitness/internal/api/render"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/web"

	"github.com/go-chi/chi"
	"github.com/swaggest/swgui/v3emb"
	"github.com/unrolled/secure"
)

// WebHandler is an abstraction of an http handler that handles web calls.
type WebHandler interface {
	http.Handler
}

// NewWebHandler returns a new WebHandler.
func NewWebHandler(config *types.Config) WebHandler {
	// Use go-chi router for inner routing
	r := chi.NewRouter()
	// create middleware to enforce security best practices for
	// the user interface. note that theis middleware is only used
	// when serving the user interface (not found handler, below).
	sec := secure.New(
		secure.Options{
			AllowedHosts:          config.Secure.AllowedHosts,
			HostsProxyHeaders:     config.Secure.HostsProxyHeaders,
			SSLRedirect:           config.Secure.SSLRedirect,
			SSLTemporaryRedirect:  config.Secure.SSLTemporaryRedirect,
			SSLHost:               config.Secure.SSLHost,
			SSLProxyHeaders:       config.Secure.SSLProxyHeaders,
			STSSeconds:            config.Secure.STSSeconds,
			STSIncludeSubdomains:  config.Secure.STSIncludeSubdomains,
			STSPreload:            config.Secure.STSPreload,
			ForceSTSHeader:        config.Secure.ForceSTSHeader,
			FrameDeny:             config.Secure.FrameDeny,
			ContentTypeNosniff:    config.Secure.ContentTypeNosniff,
			BrowserXssFilter:      config.Secure.BrowserXSSFilter,
			ContentSecurityPolicy: config.Secure.ContentSecurityPolicy,
			ReferrerPolicy:        config.Secure.ReferrerPolicy,
		},
	)

	// openapi playground endpoints
	// TODO: this should not be generated and marshaled on the fly every time?
	r.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		spec := openapi.Generate()
		data, err := spec.MarshalYAML()
		if err != nil {
			render.ErrorMessagef(w, http.StatusInternalServerError, "error serializing openapi.yaml: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
	swagger := v3emb.NewHandler("API Definition", "/openapi.yaml", "/swagger")
	r.With(sec.Handler).Handle("/swagger", swagger)
	r.With(sec.Handler).Handle("/swagger/*", swagger)

	// serve all other routes from the embedded filesystem,
	// which in turn serves the user interface.
	r.With(sec.Handler).NotFound(
		web.Handler(),
	)

	return r
}
