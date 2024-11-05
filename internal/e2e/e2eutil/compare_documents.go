// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/e2eutil/compare_documents.go
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

package e2eutil

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"reflect"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/rogpeppe/go-internal/testscript"
)

const compareDocsRequiredArgNum = 3

// **** Currently does a soft comparison, checks no nodes are lost and pkg names are the same
// Cannot compare content since cdx properties are wiped out. *********

// compareDocuments is a testscript command that compares the
// two given protobom documents and checks for equality.
func compareDocuments(script *testscript.TestScript, neg bool, args []string) {
	if len(args) != compareDocsRequiredArgNum {
		script.Fatalf("syntax: compare_docs directory_name file_name1 file_name2")
	}

	documents := []*sbom.Document{}
	sbomReader := reader.New()
	dir := args[0]

	for i := 1; i < len(args); i++ {
		fileString := path.Join(dir, args[i])

		file := getFile(script, fileString)

		data, err := io.ReadAll(file)
		if err != nil {
			script.Fatalf("failed to read from input file: %s", file.Name())
		}

		document, err := sbomReader.ParseStream(bytes.NewReader(data))
		if err != nil {
			script.Fatalf("failed to parse SBOM data from file: %s", file.Name())
		}

		documents = append(documents, document)
	}

	metaMatches := compareMetaData(script, documents[0].GetMetadata(), documents[1].GetMetadata())
	if !metaMatches {
		script.Logf("metadata does not match")
	}

	nodeListMatches := Equal(documents[0].GetNodeList(), documents[1].GetNodeList(), script)
	if !nodeListMatches {
		script.Logf("nodelist does not match")
	}

	reportResult(script, (metaMatches && nodeListMatches), neg)
}

func reportResult(script *testscript.TestScript, docsMatch, neg bool) { //nolint:revive
	if !docsMatch && !neg {
		script.Fatalf("documents do not match")
	}

	if docsMatch && neg {
		script.Fatalf("documents Match")
	}

	script.Logf("documents Match")
}

func getFile(script *testscript.TestScript, filePath string) *os.File {
	if !fileExists(filePath) {
		script.Fatalf("file not found %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		script.Fatalf("failed to open input file: %s", file.Name())
	}

	return file
}

func compareMetaData(script *testscript.TestScript, have, want *sbom.Metadata) bool {
	switch {
	case have.GetId() != want.GetId():
		script.Logf("MetaData Id does not match. have %s, want: %s", have.GetId(), want.GetId())

		return false
	case have.GetVersion() != want.GetVersion():
		script.Logf("MetaData Version does not match. have %s, want: %s", have.GetVersion(), want.GetVersion())

		return false
	case have.GetName() != want.GetName():
		script.Logf("MetaData Name does not match. have %s, want: %s", have.GetName(), want.GetName())

		return false
	case len(have.GetAuthors()) != len(want.GetAuthors()):
		script.Logf("MetaData Authors do not match. have %s, want: %s", have.GetAuthors(), want.GetAuthors())

		return false
	case len(have.GetDocumentTypes()) != len(want.GetDocumentTypes()):
		script.Logf("MetaData DocTypes do not match. have %s, want: %s",
			have.GetDocumentTypes(), want.GetDocumentTypes())

		return false
	default:
		return true
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	return !errors.Is(err, fs.ErrNotExist)
}

func Equal(nl1, nl2 *sbom.NodeList, script *testscript.TestScript) bool {
	if nl2 == nil {
		return false
	}

	// First, quick one: Compare the lengths of the internals:
	if len(nl1.GetEdges()) != len(nl2.GetEdges()) ||
		len(nl1.GetNodes()) != len(nl2.GetNodes()) ||
		len(nl1.GetRootElements()) != len(nl2.GetRootElements()) {
		script.Logf("lengths differ")

		return false
	}

	// Compare the flattened GetRootElements() list
	rel1 := nl1.GetRootElements()
	rel2 := nl2.GetRootElements()

	sort.Strings(rel1)
	sort.Strings(rel2)

	if !reflect.DeepEqual(rel1, rel2) {
		script.Logf("root elements differ")

		return false
	}

	// Compare the GetNodes()
	nlNodes := []string{}
	nl2Nodes := []string{}

	for _, n := range nl1.GetNodes() {
		nlNodes = append(nlNodes, n.GetName())
	}

	for _, n := range nl2.GetNodes() {
		nl2Nodes = append(nl2Nodes, n.GetName())
	}

	sort.Strings(nlNodes)
	sort.Strings(nl2Nodes)

	// If no logging output is seen, but equality fails
	// this means the nodes are not equal
	return cmp.Equal(nlNodes, nl2Nodes)
}
