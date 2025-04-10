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

package gitea

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/gitea/modules/git"
)

// Config set local git key and value configuration.
func (g Adapter) Config(ctx context.Context, repoPath, key, value string) error {
	var outbuf, errbuf strings.Builder
	if err := git.NewCommand(ctx, "config", "--local").AddArguments(key, value).
		Run(&git.RunOpts{
			Dir:    repoPath,
			Stdout: &outbuf,
			Stderr: &errbuf,
		}); err != nil {
		return fmt.Errorf("git config [%s -> <%s> ]: %w\n%s\n%s",
			key, value, err, outbuf.String(), errbuf.String())
	}
	return nil
}
