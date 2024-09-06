package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

const (
	aliasRemoveMinArgNum = 1
	aliasRemoveMaxArgNum = 2
	aliasSetExactArgNum  = 2
)

func aliasCmd() *cobra.Command {
	aliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Edit the alias for a document",
		Long:  "Edit the alias for a document",
	}

	aliasCmd.AddCommand(aliasRemoveCmd(), aliasSetCmd())

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

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatal(err, "documentID", args[0])
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			docAlias, err := backend.GetDocumentUniqueAnnotation(document.Metadata.Id, db.BomctlAnnotationAlias)
			if err != nil {
				backend.Logger.Fatal(err, "documentID", args[0])
			}

			if err := backend.RemoveAnnotations(document.Metadata.Id, db.BomctlAnnotationAlias, docAlias); err != nil {
				backend.Logger.Fatal(err, "name", db.BomctlAnnotationAlias, "value", docAlias)
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
		Args:  cobra.ExactArgs(aliasSetExactArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatal("Failed to get document", "documentID", args[0], "err", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			docAlias, err := backend.GetDocumentUniqueAnnotation(document.Metadata.Id, db.BomctlAnnotationAlias)
			if err != nil {
				backend.Logger.Fatal(err)
			}

			if err := backend.RemoveAnnotations(document.Metadata.Id, db.BomctlAnnotationAlias, docAlias); err != nil {
				backend.Logger.Fatal(err, db.BomctlAnnotationAlias, docAlias)
			}

			if err := backend.SetUniqueAnnotation(document.Metadata.Id,
				db.BomctlAnnotationAlias, args[1]); err != nil {
				backend.Logger.Fatal(err, db.BomctlAnnotationAlias, docAlias)
			}
		},
	}

	return aliasSetCmd
}
