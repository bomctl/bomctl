// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/import.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -----------------------------------------------------------------------------
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
// -----------------------------------------------------------------------------

package cmd

import (
	"os"
	"slices"

	"github.com/spf13/cobra"

	imprt "github.com/bomctl/bomctl/internal/pkg/import"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func importCmd() *cobra.Command {
	opts := &options.ImportOptions{}

	importCmd := &cobra.Command{
		Use:   "import [flags] { - | FILE...}",
		Args:  cobra.MinimumNArgs(1),
		Short: "Import SBOM file(s) from stdin or local filesystem",
		Long:  "Import SBOM file(s) from stdin or local filesystem",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			if slices.Contains(args, "-") && len(args) > 1 {
				opts.Logger.Fatal("Piped input and file path args cannot be specified simultaneously.")
			}

			for idx := range args {
				if args[idx] == "-" {
					opts.InputFiles = append(opts.InputFiles, os.Stdin)
				} else {
					file, err := os.Open(args[idx])
					if err != nil {
						opts.Logger.Fatal("failed to open input file", "err", err, "file", file)
					}

					opts.InputFiles = append(opts.InputFiles, file)

					defer file.Close() //nolint:revive
				}
			}

			if err := imprt.Import(opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
	}

	return importCmd
}
