// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/export.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// ------------------------------------------------------------------------
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
// ------------------------------------------------------------------------
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/export"
	"github.com/bomctl/bomctl/internal/pkg/utils"
	"github.com/bomctl/bomctl/internal/pkg/utils/format"
)

func exportCmd() *cobra.Command {
	documentIDs := []string{}
	opts := &export.ExportOptions{}

	outputFile := OutputFileValue("")
	formatString := FormatStringValue(format.DefaultFormatString())
	formatEncoding := FormatEncodingValue(format.DefaultEncoding())

	exportCmd := &cobra.Command{
		Use:   "export [flags] SBOM_URL...",
		Args:  cobra.MinimumNArgs(1),
		Short: "Export SBOM file(s) from Storage",
		Long:  "Export SBOM file(s) from Storage",
		PreRun: func(_ *cobra.Command, args []string) {
			documentIDs = append(documentIDs, args...)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			cfgFile, err := cmd.Flags().GetString("config")
			cobra.CheckErr(err)

			verbosity, err := cmd.Flags().GetCount("verbose")
			cobra.CheckErr(err)

			opts.Debug = verbosity >= minDebugLevel

			initOpts(opts, cfgFile, string(formatString), string(formatEncoding))
			backend := initBackend(opts)

			if string(outputFile) != "" {
				if len(documentIDs) > 1 {
					opts.Logger.Fatal("The --output-file option cannot be used when more than one SBOM  is provided.")
				}

				out, err := os.Create(string(outputFile))
				if err != nil {
					opts.Logger.Fatal("error creating output file", "outputFile", outputFile)
				}

				opts.OutputFile = out

				defer opts.OutputFile.Close()
			}
			Export(documentIDs, opts, backend)
		},
	}

	exportCmd.Flags().VarP(&outputFile, "output-file", "o", "Path to output file")
	exportCmd.Flags().VarP(&formatString, "format", "f", format.FormatStringOptions)
	exportCmd.Flags().VarP(&formatEncoding, "encoding", "e", "Output encoding [spdx: [text, json] cyclonedx: [json]")

	return exportCmd
}

func Export(documentIDs []string, opts *export.ExportOptions, backend *db.Backend) {
	for _, id := range documentIDs {
		if err := export.Export(id, opts, backend); err != nil {
			opts.Logger.Fatal(err)
		}
	}
}

func initOpts(opts *export.ExportOptions, cfgFile, formatString, formatEncoding string) {
	opts.CacheDir = viper.GetString("cache_dir")
	opts.ConfigFile = cfgFile
	opts.FormatString = formatString
	opts.Encoding = formatEncoding
}

func initBackend(opts *export.ExportOptions) *db.Backend {
	backend := db.NewBackend(func(b *db.Backend) {
		b.Options.DatabaseFile = filepath.Join(opts.CacheDir, db.DatabaseFile)
		b.Options.Debug = opts.Debug
		b.Logger = utils.NewLogger("export")
	})

	if err := backend.InitClient(); err != nil {
		backend.Logger.Fatalf("failed to initialize backend client: %v", err)
	}

	return backend
}
