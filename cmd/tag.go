package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	tagAddArgNum       int = 2
	tagClearArgNum     int = 1
	tagListArgNum      int = 1
	tagRemoveMinArgNum int = 2
)

func tagCmd() *cobra.Command {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Edit the tags of a document",
		Long:  "Edit the tags of a document",
	}

	tagCmd.AddCommand(tagClearCmd(), tagAddCmd(), tagRemoveCmd(), tagListCmd())

	return tagCmd
}

func tagAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [flags] SBOM_ID TAGS...",
		Short: "Add tags to a document",
		Long:  "Add tags to a document",
		Args:  cobra.MinimumNArgs(tagAddArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByID(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			if err := backend.AddAnnotations(document.Metadata.Id, "tag", args[1:]...); err != nil {
				backend.Logger.Fatalf("failed to add tags: %v", err)
			}
		},
	}

	return addCmd
}

func tagClearCmd() *cobra.Command {
	clearCmd := &cobra.Command{
		Use:   "clear [flags] SBOM_ID...",
		Short: "Clear the tags of a document",
		Long:  "Clear the tags of a document",
		Args:  cobra.ExactArgs(tagClearArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByID(args[0])
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

func tagListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags] SBOM_ID",
		Aliases: []string{"ls"},
		Short:   "List the tags of a document",
		Long:    "List the tags of a document",
		Args:    cobra.ExactArgs(tagListArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByID(args[0])
			if err != nil {
				backend.Logger.Fatal("Failed to get document", "err", err)
			}

			annotations, err := backend.GetDocumentAnnotations(document.Metadata.Id, "tag")
			if err != nil {
				backend.Logger.Fatal("Failed to get document tags", "err", err)
			}

			for _, annotation := range annotations {
				fmt.Fprintln(os.Stdout, annotation.Value)
			}
		},
	}

	return listCmd
}

func tagRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] SBOM_ID",
		Aliases: []string{"rm"},
		Short:   "Remove the tags of a document",
		Long:    "Remove the tags of a document",
		Args:    cobra.MinimumNArgs(tagRemoveMinArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByID(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			err = backend.RemoveAnnotations(document.Metadata.Id, "tag", args[1:]...) //nolint:revive
			if err != nil {
				backend.Logger.Fatalf("failed to remove tags: %v", err)
			}
		},
	}

	return removeCmd
}
