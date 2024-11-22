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
	"path"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/rogpeppe/go-internal/testscript"
	"google.golang.org/protobuf/proto"
)

const compareDocsRequiredArgNum = 3

// **** Currently does a soft comparison, checks no nodes are lost and pkg names are the same
// Cannot compare content since cdx properties are wiped out. *********

// compareDocuments is a testscript command that compares the
// two given protobom documents and checks for equality.
// **** Currently does a soft comparison, checks no nodes are lost and pkg names are the same
// Cannot compare content since cdx properties are wiped out. *********.
func compareDocuments(script *testscript.TestScript, neg bool, args []string) {
	if len(args) != compareDocsRequiredArgNum {
		script.Fatalf("syntax: compare_docs directory_name file_name1 file_name2")
	}

	documents := []*sbom.Document{}
	sbomReader := reader.New()
	dir := args[0]

	for i := 1; i < len(args); i++ {
		fileString := path.Join(dir, args[i])

		file := script.ReadFile(fileString)

		document, err := sbomReader.ParseStream(strings.NewReader(file))
		script.Check(err)

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

func reportResult(script *testscript.TestScript, docsMatch, neg bool) { //revive:disable:flag-parameter
	if !docsMatch && !neg {
		script.Fatalf("documents do not match")
	}

	if docsMatch && neg {
		script.Fatalf("documents match")
	}

	script.Logf("document comparison passed")
}

// checks for metadata equality, taking expected diffs into account.
func compareMetaData(script *testscript.TestScript, have, want *sbom.Metadata) bool {
	// append protobom tool entry to match expected value
	if format := want.GetSourceData().GetFormat(); strings.Contains(format, "spdx") {
		want.Tools = append(want.GetTools(), &sbom.Tool{Name: "protobom-devel"})
		slices.SortFunc(want.GetTools(), func(i, j *sbom.Tool) int {
			return strings.Compare(i.GetName(), j.GetName())
		})
	}

	// Remove SourceData to avoid hash mismatches due to format specific losses.
	have.SourceData = nil
	want.SourceData = nil

	// Remove Date to ignore time created mismatch.
	have.Date = nil
	want.Date = nil

	haveBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(have)
	if err != nil {
		script.Fatalf("failed to marshal metadata: %v", err)
	}

	wantBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(want)
	if err != nil {
		script.Fatalf("failed to marshal metadata: %v", err)
	}

	return bytes.Equal(haveBytes, wantBytes)
}

// Performs soft comparison of two nodelists.
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
