// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/push.go
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
package cmd

import (
	"github.com/protobom/protobom/pkg/formats"
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/client"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/push"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

const pushArgNum = 2

func pushCmd() *cobra.Command {
	opts := &client.PushOptions{
		Options: options.New(options.WithLogger(utils.NewLogger("push"))),
	}

	pushCmd := &cobra.Command{
		Use:    "push [flags] SBOM_ID [SBOM_ID...] DEST_PATH",
		Args:   cobra.MinimumNArgs(pushArgNum),
		Short:  "Push stored SBOM file(s) to remote URL or filesystem",
		Long:   "Push stored SBOM file(s) to remote URL or filesystem",
		PreRun: preRun(opts.Options),
		Run: func(_ *cobra.Command, args []string) {
			if err := push.Push(args[:len(args)-1], args[len(args)-1], opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
	}

	pushCmd.Flags().StringP("encoding", "e", formats.JSON, encodingHelp())
	pushCmd.Flags().StringP("format", "f", formats.CDXFORMAT, formatHelp())
	pushCmd.Flags().BoolVar(&opts.UseNetRC, "netrc", false, "Use .netrc file for authentication to remote hosts")
	pushCmd.Flags().BoolVar(&opts.UseTree, "tree", false, "Recursively push all SBOMs in external reference tree")

	return pushCmd
}
