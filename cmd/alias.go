package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/options"
)

type AliasOptions struct {
	*options.Options
	UseAlias bool
}

const (
	aliasRemoveMinArgNum = 1
	aliasRemoveMaxArgNum = 2
	aliasSetArgNum       = 2
)

func aliasCmd() *cobra.Command {
	aliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Edit the alias for a document",
		Long:  "Edit the alias for a document",
	}

	aliasCmd.AddCommand(aliasSetCmd())
	aliasCmd.AddCommand(aliasRemoveCmd())

	return aliasCmd
}

func aliasRemoveCmd() *cobra.Command {
	aliasCmd := &cobra.Command{
		Use:     "remove [flags] SBOM_ID",
		Aliases: []string{"rm"},
		Short:   "Remove the alias for a specific document",
		Long:    "Remove the alias for a specific document",
		Args:    cobra.RangeArgs(aliasRemoveMinArgNum, aliasRemoveMaxArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatal(err, "documentID", args[0])
			}

			docAlias, err := backend.GetDocumentAlias(document.Metadata.Id)
			if err != nil {
				backend.Logger.Fatal(err, "documentID", args[0])
			}

			if err := backend.RemoveAnnotations(document.Metadata.Id, "tag", docAlias); err != nil {
				backend.Logger.Fatal(err, "alias", docAlias)
			}
		},
	}

	return aliasCmd
}

func aliasSetCmd() *cobra.Command {
	aliasSetCmd := &cobra.Command{
		Use:   "set [flags] SBOM_ID NEW_ALIAS",
		Short: "Set the alias for a specific document",
		Long:  "Set the alias for a specific document",
		Args:  cobra.ExactArgs(aliasSetArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatal("Failed to get document", "documentID", args[0], "err", err)
			}

			docAlias, err := backend.GetDocumentAlias(document.Metadata.Id)
			if err != nil {
				backend.Logger.Fatal(err)
			}

			if err := backend.RemoveAnnotations(document.Metadata.Id, "alias", docAlias); err != nil {
				backend.Logger.Fatal(err, "alias", docAlias)
			}

			if len(args) > 1 {
				if err := backend.AddAnnotations(document.Metadata.Id, "alias", args[1]); err != nil {
					backend.Logger.Fatal(err, "alias", docAlias)
				}
			}
		},
	}

	return aliasSetCmd
}
