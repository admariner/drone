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

package authn

import (
	"errors"
	"net/http"

	"github.com/harness/gitness/internal/auth"
)

var (
	// ErrNoAuthData that is returned if the authorizer doesn't find any data in the request that can be used for auth.
	ErrNoAuthData = errors.New("the request doesn't contain any auth data that can be used by the Authorizer")
	// ErrNotAcceptedAuthData that is returned if the request is using an auth data that is not accepted by the authorizer.
	// e.g, don't accept jwt (without allowedResources field) for git clone/pull request.
	ErrNotAcceptedAuthMethod = errors.New("the request contains auth method that is not accepted by the Authorizer")
)

type SourceRouter string

const (
	SourceRouterAPI SourceRouter = "api"
	SourceRouterGIT SourceRouter = "git"
)

// Authenticator is an abstraction of an entity that's responsible for authenticating principals
// that are making calls via HTTP.
type Authenticator interface {
	/*
	 * Tries to authenticate the acting principal if credentials are available.
	 * Returns:
	 *		(session, nil) 		    - request contains auth data and principal was verified
	 *		(nil, ErrNoAuthData)	- request doesn't contain any auth data
	 *		(nil, err)  			- request contains auth data but verification failed
	 */
	Authenticate(r *http.Request, sourceRouter SourceRouter) (*auth.Session, error)
}
