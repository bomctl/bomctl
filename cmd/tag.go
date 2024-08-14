package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

type TagOptions struct {
	*options.Options
}

func tagCmd() *cobra.Command {

	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Edit the tags of a document",
		Long:  "Edit the tags of a document",
	}

	tagCmd.AddCommand(tagClearCmd(), tagAddCmd(), tagRemoveCmd(), tagListCmd())

	return tagCmd
}

func tagClearCmd() *cobra.Command {

	clearCmd := &cobra.Command{
		Use:   "clear [flags] SBOM_ID...",
		Short: "Clear the tags of a document",
		Long:  "Clear the tags of a document",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbosity, err := cmd.Flags().GetCount("verbose")
			cobra.CheckErr(err)

			backend := db.NewBackend().
				Debug(verbosity >= minDebugLevel).
				WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)).
				WithLogger(utils.NewLogger("tag"))

			if err := backend.InitClient(); err != nil {
				backend.Logger.Fatalf("failed to initialize backend client: %v", err)
			}

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			annotationsToRemove, err := backend.GetDocumentAnnotations(document.Metadata.Id, "tag")
			if err != nil {
				backend.Logger.Fatalf("failed to clear tags: %v", err)
			}

			tagsToRemove := []string{}
			for _, annotation := range annotationsToRemove {
				tagsToRemove = append(tagsToRemove, annotation.Value)
			}

			err = backend.RemoveAnnotations(document.Metadata.Id, "tag", tagsToRemove...)
			if err != nil {
				backend.Logger.Fatalf("failed to clear tags: %v", err)
			}
		},
	}

	return clearCmd
}

func tagAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [flags] SBOM_ID TAGS...",
		Short: "Add tags to a document",
		Long:  "Add tags to a document",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			verbosity, err := cmd.Flags().GetCount("verbose")
			cobra.CheckErr(err)

			backend := db.NewBackend().
				Debug(verbosity >= minDebugLevel).
				WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)).
				WithLogger(utils.NewLogger("tag"))

			if err := backend.InitClient(); err != nil {
				backend.Logger.Fatalf("failed to initialize backend client: %v", err)
			}

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			backend.AddAnnotations(document.Metadata.Id, "tag", args[1:]...)
		},
	}

	return addCmd
}

func tagRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] SBOM_ID",
		Aliases: []string{"rm"},
		Short:   "Remove the tags of a document",
		Long:    "Remove the tags of a document",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbosity, err := cmd.Flags().GetCount("verbose")
			cobra.CheckErr(err)

			backend := db.NewBackend().
				Debug(verbosity >= minDebugLevel).
				WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)).
				WithLogger(utils.NewLogger("tag"))

			if err := backend.InitClient(); err != nil {
				backend.Logger.Fatalf("failed to initialize backend client: %v", err)
			}

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			err = backend.RemoveAnnotations(document.Metadata.Id, "tag", args[1:]...)
			if err != nil {
				backend.Logger.Fatalf("failed to remove tags: %v", err)
			}
		},
	}

	return removeCmd
}

func tagListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags] SBOM_ID",
		Aliases: []string{"ls"},
		Short:   "List the tags of a document",
		Long:    "List the tags of a document",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbosity, err := cmd.Flags().GetCount("verbose")
			cobra.CheckErr(err)

			backend := db.NewBackend().
				Debug(verbosity >= minDebugLevel).
				WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)).
				WithLogger(utils.NewLogger("tag"))

			if err := backend.InitClient(); err != nil {
				backend.Logger.Fatalf("failed to initialize backend client: %v", err)
			}

			defer backend.CloseClient()

			document, err := backend.GetDocument(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			annotations, err := backend.GetDocumentAnnotations(document.Metadata.Id, "tag")
			if err != nil {
				backend.Logger.Fatalf("failed to get document tags: %v", err)
			}

			for _, annotation := range annotations {
				fmt.Println(annotation.Value)
			}
		},
	}

	return listCmd
}
