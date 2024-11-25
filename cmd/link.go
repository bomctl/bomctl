// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/link.go
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
	"strings"

	"github.com/spf13/cobra"
)

func linkCmd() *cobra.Command {
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Edit links between documents and/or nodes",
		Long:  "Edit links between documents and/or nodes",
	}

	linkCmd.AddCommand(linkAddCmd(), linkClearCmd(), linkListCmd(), linkRemoveCmd())

	return linkCmd
}

func linkAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [flags] {node:NODE_ID | document:SBOM_ID} [document:]SBOM_ID",
		Short: "Add a link from a document or node to a document",
		Long:  "Add a link from a document or node to a document",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opts := optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			if !(strings.HasPrefix(args[0], "document:") || strings.HasPrefix(args[0], "node:")) {
				opts.Logger.Fatal(
					"Must be prefixed with either 'document:' or 'node:'", "argument", args[0],
				)
			}

			opts.Logger.Warn("Not yet implemented.")
		},
	}

	return addCmd
}

func linkClearCmd() *cobra.Command {
	clearCmd := &cobra.Command{
		Use:   "clear [flags] {node:NODE_ID | document:SBOM_ID}...",
		Short: "Remove all links from specified documents and nodes",
		Long:  "Remove all links from specified documents and nodes",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts := optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			for _, arg := range args {
				if !(strings.HasPrefix(arg, "document:") || strings.HasPrefix(arg, "node:")) {
					opts.Logger.Fatal(
						"Positional arguments must be prefixed with either 'document:' or 'node:'", "argument", arg,
					)
				}
			}

			opts.Logger.Warn("Not yet implemented.")
		},
	}

	return clearCmd
}

func linkListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags] {node:NODE_ID | document:SBOM_ID}",
		Aliases: []string{"ls"},
		Short:   "List the links of a document or node",
		Long:    "List the links of a document or node",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts := optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			if !(strings.HasPrefix(args[0], "document:") || strings.HasPrefix(args[0], "node:")) {
				opts.Logger.Fatal(
					"Must be prefixed with either 'document:' or 'node:'", "argument", args[0],
				)
			}

			opts.Logger.Warn("Not yet implemented.")
		},
	}

	return listCmd
}

func linkRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] {node:NODE_ID | document:SBOM_ID} [document:]SBOM_ID...",
		Aliases: []string{"rm"},
		Short:   "Remove specified links from a document or node",
		Long:    "Remove specified links from a document or node",
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opts := optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			if !(strings.HasPrefix(args[0], "document:") || strings.HasPrefix(args[0], "node:")) {
				opts.Logger.Fatal(
					"Must be prefixed with either 'document:' or 'node:'", "argument", args[0],
				)
			}

			opts.Logger.Warn("Not yet implemented.")
		},
	}

	return removeCmd
}
