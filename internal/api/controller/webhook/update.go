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
	"context"
	"fmt"

	"github.com/harness/gitness/internal/auth"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/check"
	"github.com/harness/gitness/types/enum"
)

type UpdateInput struct {
	DisplayName *string               `json:"display_name"`
	Description *string               `json:"description"`
	URL         *string               `json:"url"`
	Secret      *string               `json:"secret"`
	Enabled     *bool                 `json:"enabled"`
	Insecure    *bool                 `json:"insecure"`
	Triggers    []enum.WebhookTrigger `json:"triggers"`
}

// Update updates an existing webhook.
func (c *Controller) Update(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	webhookID int64,
	in *UpdateInput,
) (*types.Webhook, error) {
	repo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoEdit)
	if err != nil {
		return nil, err
	}

	// get the hook and ensure it belongs to us
	hook, err := c.getWebhookVerifyOwnership(ctx, repo.ID, webhookID)
	if err != nil {
		return nil, err
	}

	// validate input
	if err = checkUpdateInput(in, c.allowLoopback, c.allowPrivateNetwork); err != nil {
		return nil, err
	}

	// update webhook struct (only for values that are provided)
	if in.DisplayName != nil {
		hook.DisplayName = *in.DisplayName
	}
	if in.Description != nil {
		hook.Description = *in.Description
	}
	if in.URL != nil {
		hook.URL = *in.URL
	}
	if in.Secret != nil {
		encryptedSecret, err := c.encrypter.Encrypt(*in.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt webhook secret: %w", err)
		}
		hook.Secret = string(encryptedSecret)
	}
	if in.Enabled != nil {
		hook.Enabled = *in.Enabled
	}
	if in.Insecure != nil {
		hook.Insecure = *in.Insecure
	}
	if in.Triggers != nil {
		hook.Triggers = deduplicateTriggers(in.Triggers)
	}

	if err = c.webhookStore.Update(ctx, hook); err != nil {
		return nil, err
	}

	return hook, nil
}

func checkUpdateInput(in *UpdateInput, allowLoopback bool, allowPrivateNetwork bool) error {
	if in.DisplayName != nil {
		if err := check.DisplayName(*in.DisplayName); err != nil {
			return err
		}
	}
	if in.Description != nil {
		if err := check.Description(*in.Description); err != nil {
			return err
		}
	}
	if in.URL != nil {
		if err := checkURL(*in.URL, allowLoopback, allowPrivateNetwork); err != nil {
			return err
		}
	}
	if in.Secret != nil {
		if err := checkSecret(*in.Secret); err != nil {
			return err
		}
	}
	if in.Triggers != nil {
		if err := checkTriggers(in.Triggers); err != nil {
			return err
		}
	}

	return nil
}
