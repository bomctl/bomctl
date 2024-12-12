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
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/link"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func linkCmd() *cobra.Command {
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Edit links between documents and/or nodes",
		Long:  "Edit links between documents and/or nodes",
	}

	typeChoice := newChoiceValue("Type referenced by SRC_ID", "node", "document")
	linkCmd.PersistentFlags().VarP(typeChoice, "type", "t", typeChoice.Usage())

	linkCmd.AddCommand(linkAddCmd(), linkClearCmd(), linkListCmd(), linkRemoveCmd())

	return linkCmd
}

func linkAddCmd() *cobra.Command {
	opts := &options.LinkOptions{}

	addCmd := &cobra.Command{
		Use:   "add [flags] SRC_ID SBOM_ID",
		Short: "Add a link from a document or node to a document",
		Long:  "Add a link from a document or node to a document",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			parseLinkArgs(cmd, args[:1], opts)

			opts.ToIDs = append(opts.ToIDs, args[1])

			if err := link.AddLink(backend, opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
	}

	return addCmd
}

func linkClearCmd() *cobra.Command {
	opts := &options.LinkOptions{}

	clearCmd := &cobra.Command{
		Use:   "clear [flags] SRC_ID...",
		Short: "Remove all links from specified documents and nodes",
		Long:  "Remove all links from specified documents and nodes",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			parseLinkArgs(cmd, args, opts)

			if err := link.ClearLinks(backend, opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
	}

	return clearCmd
}

func linkListCmd() *cobra.Command {
	opts := &options.LinkOptions{}

	listCmd := &cobra.Command{
		Use:     "list [flags] SRC_ID",
		Aliases: []string{"ls"},
		Short:   "List the links of a document or node",
		Long:    "List the links of a document or node",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			parseLinkArgs(cmd, args, opts)

			if err := link.ListLinks(backend, opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
	}

	return listCmd
}

func linkRemoveCmd() *cobra.Command {
	opts := &options.LinkOptions{}

	removeCmd := &cobra.Command{
		Use:     "remove [flags] SRC_ID SBOM_ID...",
		Aliases: []string{"rm"},
		Short:   "Remove specified links from a document or node",
		Long:    "Remove specified links from a document or node",
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			parseLinkArgs(cmd, args[:1], opts)

			opts.ToIDs = append(opts.ToIDs, args[1:]...)

			if err := link.RemoveLink(backend, opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
	}

	return removeCmd
}

func parseLinkArgs(cmd *cobra.Command, args []string, opts *options.LinkOptions) {
	switch cmd.Flag("type").Value.String() {
	case "document":
		opts.DocumentIDs = append(opts.DocumentIDs, args...)
	case "node":
		opts.NodeIDs = append(opts.NodeIDs, args...)
	}
}
