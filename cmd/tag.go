package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

const (
	tagAddMinArgs     int = 2
	tagClearExactArgs int = 1
	tagListExactArgs  int = 1
	tagRemoveMinArgs  int = 2
)

func tagCmd() *cobra.Command {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Edit the tags of a document",
		Long:  "Edit the tags of a document",
	}

	tagCmd.AddCommand(tagAddCmd(), tagClearCmd(), tagListCmd(), tagRemoveCmd())

	return tagCmd
}

func tagAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [flags] SBOM_ID TAGS...",
		Short: "Add tags to a document",
		Long:  "Add tags to a document",
		Args:  cobra.MinimumNArgs(tagAddMinArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			if err := backend.AddAnnotations(document.Metadata.Id, db.BomctlAnnotationTag, args[1:]...); err != nil {
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
		Args:  cobra.ExactArgs(tagClearExactArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			annotationsToRemove, err := backend.GetDocumentAnnotations(document.Metadata.Id, db.BomctlAnnotationTag)
			if err != nil {
				backend.Logger.Fatalf("failed to clear tags: %v", err)
			}

			tagsToRemove := make([]string, len(annotationsToRemove))
			for idx := range annotationsToRemove {
				tagsToRemove[idx] = annotationsToRemove[idx].Value
			}

			err = backend.RemoveAnnotations(document.Metadata.Id, db.BomctlAnnotationTag, tagsToRemove...)
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
		Args:    cobra.ExactArgs(tagListExactArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatal("Failed to get document", "err", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			annotations, err := backend.GetDocumentAnnotations(document.Metadata.Id, db.BomctlAnnotationTag)
			if err != nil {
				backend.Logger.Fatal("Failed to get document tags", "err", err)
			}

			sortedAnnotations := make([]string, len(annotations))

			for idx := range annotations {
				sortedAnnotations[idx] = annotations[idx].Value
			}

			sort.Strings(sortedAnnotations)

			fmt.Printf("\nTags for %v\n%v\n", args[0], strings.Repeat("─", 80))
			fmt.Printf("%v\n\n", strings.Join(sortedAnnotations, "\n"))
		},
	}

	return listCmd
}

func tagRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] SBOM_ID TAGS...",
		Aliases: []string{"rm"},
		Short:   "Remove the tags of a document",
		Long:    "Remove the tags of a document",
		Args:    cobra.MinimumNArgs(tagRemoveMinArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			err = backend.RemoveAnnotations(document.Metadata.Id, db.BomctlAnnotationTag, args[1:]...)
			if err != nil {
				backend.Logger.Fatalf("failed to remove tags: %v", err)
			}
		},
	}

	return removeCmd
}
