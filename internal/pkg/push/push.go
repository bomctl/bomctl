// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/push/push.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// ------------------------------------------------------------------------
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ------------------------------------------------------------------------
package push

import (
	"fmt"

	"github.com/bomctl/bomctl/internal/pkg/client"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func Push(sbomID, destPath string, opts *options.PushOptions) error {
	opts.Logger.Info("Pushing Document", "sbomID", sbomID)

	pusher, err := client.New(destPath)
	if err != nil {
		return fmt.Errorf("creating push client: %w", err)
	}

	opts.Logger.Info(fmt.Sprintf("Pushing to %s URL", pusher.Name()), "url", destPath)

	err = pusher.Push(sbomID, destPath, opts)
	if err != nil {
		return fmt.Errorf("failed to push to %s: %w", destPath, err)
	}

	return nil
}
