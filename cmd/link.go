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
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/link"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

const (
	blue    = lipgloss.ANSIColor(termenv.ANSIBlue)
	cyan    = lipgloss.ANSIColor(termenv.ANSICyan)
	green   = lipgloss.ANSIColor(termenv.ANSIGreen)
	magenta = lipgloss.ANSIColor(termenv.ANSIMagenta)
	yellow  = lipgloss.ANSIColor(termenv.ANSIYellow)
)

func linkCmd() *cobra.Command {
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Edit links between documents and/or nodes",
		Long:  "Edit links between documents and/or nodes",
	}

	typeValue := newChoiceValue("Type referenced by SRC_ID", "node", "document")
	linkCmd.PersistentFlags().VarP(typeValue, "type", "t", typeValue.Usage())

	cobra.CheckErr(linkCmd.RegisterFlagCompletionFunc("type", typeValue.CompletionFunc()))

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

			opts.Links = populateLinkTargets(cmd.Flag("type").Value.String(), args[:1], args[1:], backend)

			if err := link.AddLink(backend, opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
		ValidArgsFunction: completions,
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

			opts.Links = populateLinkTargets(cmd.Flag("type").Value.String(), args, nil, backend)

			if err := link.ClearLinks(backend, opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
		ValidArgsFunction: completions,
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

			opts.Links = populateLinkTargets(cmd.Flag("type").Value.String(), args[:1], nil, backend)

			incoming, err := link.ListLinks(backend, opts)
			if err != nil {
				opts.Logger.Fatal(err)
			}

			fmt.Fprintln(os.Stdout, newLinksTree(opts.Links[0], incoming))
		},
		ValidArgsFunction: completions,
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

			opts.Links = populateLinkTargets(cmd.Flag("type").Value.String(), args[:1], args[1:], backend)

			if err := link.RemoveLink(backend, opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
		ValidArgsFunction: completions,
	}

	return removeCmd
}

func newLinksTree(links options.Link, incoming []options.LinkTarget) *tree.Tree {
	style := lipgloss.NewStyle()

	hasOutgoing, hasIncoming := len(links.To) > 0, len(incoming) > 0

	outgoingTree := tree.Root("Outgoing links:").
		Child(sliceutil.Extract(links.To, func(lt options.LinkTarget) string { return lt.String() })).
		ItemStyle(style.Foreground(blue)).
		Hide(!hasOutgoing)

	incomingTree := tree.Root("Incoming links:").
		Child(sliceutil.Extract(incoming, func(lt options.LinkTarget) string { return lt.String() })).
		ItemStyleFunc(func(_ tree.Children, idx int) lipgloss.Style {
			if hasIncoming && incoming[idx].Type == options.LinkTargetTypeNode {
				return style.Foreground(yellow)
			}

			return style.Foreground(blue)
		}).
		Hide(!hasIncoming)

	linksTree := tree.Root(fmt.Sprintf("Links for %s:", style.Foreground(cyan).Render(links.From.String()))).
		Indenter(func(_ tree.Children, _ int) string { return " " }).
		Enumerator(func(_ tree.Children, _ int) string { return " " }).
		Child(
			tree.Root(" ").Hide(!hasOutgoing),
			outgoingTree,
			tree.Root(" ").Hide(!hasIncoming),
			incomingTree,
		)

	return linksTree
}

func populateLinkTargets(linkType string, from, to []string, backend *db.Backend) []options.Link {
	links := []options.Link{}
	targets := []options.LinkTarget{}

	for _, arg := range to {
		target := options.LinkTarget{
			Alias: arg,
			ID:    resolveDocumentID(arg, backend),
			Type:  options.LinkTargetTypeDocument,
		}

		if target.Alias == target.ID {
			target.Alias = backend.GetDocumentAlias(target.ID)
		}

		targets = append(targets, target)
	}

	switch linkType {
	case "document":
		for _, arg := range from {
			target := options.LinkTarget{
				Alias: arg,
				ID:    resolveDocumentID(arg, backend),
				Type:  options.LinkTargetTypeDocument,
			}

			if target.Alias == target.ID {
				target.Alias = backend.GetDocumentAlias(target.ID)
			}

			links = append(links, options.Link{From: target, To: targets})
		}
	case "node":
		for _, arg := range from {
			links = append(links, options.Link{
				From: options.LinkTarget{
					ID:   arg,
					Type: options.LinkTargetTypeNode,
				},
				To: targets,
			})
		}
	}

	return links
}

func resolveDocumentID(id string, backend *db.Backend) string {
	document, err := backend.GetDocumentByIDOrAlias(id)
	cobra.CheckErr(err)

	if document == nil {
		backend.Logger.Warn("Document not found", "id", id)
	}

	return document.GetMetadata().GetId()
}
