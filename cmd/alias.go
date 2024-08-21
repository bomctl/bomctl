package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

type AliasOptions struct {
	*options.Options
	UseAlias bool
}

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

func aliasSetCmd() *cobra.Command {
	aliasSetCmd := &cobra.Command{
		Use:   "set [flags] SBOM_ID NEW_ALIAS",
		Short: "Set the alias for a specific document",
		Long:  "Set the alias for a specific document",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			backend, err := db.NewBackend(
				db.WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)))

			if err := backend.InitClient(); err != nil {
				backend.Logger.Fatal("Failed to initialize backend client", "err", err)
			}

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatal("Failed to get document", "documentID", args[0], "err", err)
			}

			docAlias, err := backend.GetDocumentAlias(document.Metadata.Id)
			if err != nil {
				backend.Logger.Fatal("Failed to read alias", "err", err)
			}

			if err := backend.RemoveAnnotations(document.Metadata.Id, "alias", docAlias); err != nil {
				backend.Logger.Fatal("Failed to remove alias", "alias", docAlias, "err", err)
			}

			if len(args) > 1 {
				if err := backend.AddAnnotations(document.Metadata.Id, "alias", args[1]); err != nil {
					backend.Logger.Fatal("Failed to set alias", "alias", docAlias, "err", err)
				}
			}
		},
	}

	return aliasSetCmd
}

func aliasRemoveCmd() *cobra.Command {
	aliasCmd := &cobra.Command{
		Use:     "remove [flags] SBOM_ID",
		Aliases: []string{"rm"},
		Short:   "Remove the alias for a specific document",
		Long:    "Remove the alias for a specific document",
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			backend, err := db.NewBackend(
				db.WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)))

			if err := backend.InitClient(); err != nil {
				backend.Logger.Fatalf("failed to initialize backend client: %v", err)
			}

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			docAlias, err := backend.GetDocumentAlias(document.Metadata.Id)
			if err != nil {
				backend.Logger.Fatalf("failed to read alias: %v", err)
			}

			if err := backend.RemoveAnnotations(document.Metadata.Id, "alias", docAlias); err != nil {
				backend.Logger.Fatalf("failed to remove alias: %v", err)
			}
		},
	}

	return aliasCmd
}
