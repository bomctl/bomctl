// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
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
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/native"
	"github.com/protobom/protobom/pkg/native/serializers"
	"github.com/protobom/protobom/pkg/writer"
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/export"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

var (
	errEncodingNotSupported = errors.New("encoding not supported for selected format")
	errFormatNotSupported   = errors.New("format not supported")
)

func exportCmd() *cobra.Command {
	opts := &options.ExportOptions{}
	outputFile := outputFileValue("")

	exportCmd := &cobra.Command{
		Use:   "export [flags] SBOM_ID...",
		Args:  cobra.MinimumNArgs(1),
		Short: "Export stored SBOM(s) to filesystem",
		Long:  "Export stored SBOM(s) to filesystem",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			formatString, err := cmd.Flags().GetString("format")
			cobra.CheckErr(err)

			encoding, err := cmd.Flags().GetString("encoding")
			cobra.CheckErr(err)

			format, err := parseFormat(formatString, encoding)
			if err != nil {
				opts.Logger.Fatal(err, "format", formatString, "encoding", encoding)
			}

			opts.Format = format

			if outputFile != "" {
				if len(args) > 1 {
					opts.Logger.Fatal("The --output-file option cannot be used when more than one SBOM is provided.")
				}

				out, err := os.Create(outputFile.String())
				if err != nil {
					opts.Logger.Fatal("error creating output file", "outputFile", outputFile)
				}

				opts.OutputFile = out

				defer opts.OutputFile.Close()
			}

			// Get the documents to obtain their IDs, in case the provided IDs were aliases.
			documents, err := backend.GetDocumentsByIDOrAlias(args...)
			if err != nil {
				opts.Logger.Fatal(err, "documentIDs", args)
			}

			for _, document := range documents {
				if err := export.Export(document.GetMetadata().GetId(), opts); err != nil {
					opts.Logger.Fatal(err)
				}
			}
		},
	}

	exportCmd.Flags().VarP(&outputFile, "output-file", "o", "path to output file")
	exportCmd.Flags().StringP("format", "f", formats.CDXFORMAT, formatHelp())
	exportCmd.Flags().StringP("encoding", "e", formats.JSON, encodingHelp())

	return exportCmd
}

func encodingHelp() string {
	return fmt.Sprintf("output encoding [%s: [%s], %s: [%s]]",
		formats.SPDXFORMAT, strings.Join(encodingOptions()[formats.SPDXFORMAT], ", "),
		formats.CDXFORMAT, strings.Join(encodingOptions()[formats.CDXFORMAT], ", "),
	)
}

func encodingOptions() map[string][]string {
	return map[string][]string{
		formats.CDXFORMAT:  {formats.JSON, formats.XML},
		formats.SPDXFORMAT: {formats.JSON},
	}
}

func formatHelp() string {
	return fmt.Sprintf("SBOM output format [%s]", strings.Join(formatOptions(), ", "))
}

func formatOptions() []string {
	spdxFormats := []string{
		formats.SPDXFORMAT,
		formats.SPDXFORMAT + "-2.3",
	}

	cdxFormats := []string{
		formats.CDXFORMAT,
		formats.CDXFORMAT + "-1.0",
		formats.CDXFORMAT + "-1.1",
		formats.CDXFORMAT + "-1.2",
		formats.CDXFORMAT + "-1.3",
		formats.CDXFORMAT + "-1.4",
		formats.CDXFORMAT + "-1.5",
	}

	return append(spdxFormats, cdxFormats...)
}

func parseFormat(formatStr, encoding string) (formats.Format, error) {
	results := map[string]string{}
	pattern := regexp.MustCompile("^(?P<name>[^-]+)(?:-(?P<version>.*))?")
	match := pattern.FindStringSubmatch(formatStr)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	baseFormat := results["name"]
	version := results["version"]

	if err := validateFormat(baseFormat); err != nil {
		return formats.EmptyFormat, err
	}

	if err := validateEncoding(formatStr, encoding); err != nil {
		return formats.EmptyFormat, err
	}

	var serializer native.Serializer

	switch baseFormat {
	case formats.CDXFORMAT:
		if version == "" {
			version = "1.5"
		}

		baseFormat = "application/vnd.cyclonedx"
		serializer = serializers.NewCDX(version, encoding)
	case formats.SPDXFORMAT:
		if version == "" {
			version = "2.3"
		}

		baseFormat = "text/spdx"
		serializer = serializers.NewSPDX23()
	}

	format := formats.Format(fmt.Sprintf("%s+%s;version=%s", baseFormat, encoding, version))
	writer.RegisterSerializer(format, serializer)

	return format, nil
}

func validateEncoding(fs, encoding string) error {
	if !slices.Contains(encodingOptions()[fs], encoding) {
		return errEncodingNotSupported
	}

	return nil
}

func validateFormat(format string) error {
	if !slices.Contains(formatOptions(), format) {
		return errFormatNotSupported
	}

	return nil
}
