// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/utils/utils.go
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
package utils

import (
	"bytes"

	"github.com/bom-squad/protobom/pkg/reader"
	"github.com/bom-squad/protobom/pkg/sbom"
)

var sbomReader = reader.New()

// Return all ExternalReferences of type "BOM" in an SBOM document.
func GetBOMReferences(document *sbom.Document) (refs []*sbom.ExternalReference) {
	for _, node := range document.NodeList.Nodes {
		for _, ref := range node.GetExternalReferences() {
			if ref.Type == sbom.ExternalReference_BOM {
				refs = append(refs, ref)
			}
		}
	}

	return
}

// Parse raw byte content and return SBOM document.
func ParseSBOMData(data []byte) (document *sbom.Document, err error) {
	bytesReader := bytes.NewReader(data)
	document, err = sbomReader.ParseStream(bytesReader)

	return
}

// Parse local file and return SBOM document.
func ParseSBOMFile(filepath string) (document *sbom.Document, err error) {
	document, err = sbomReader.ParseFile(filepath)

	return
}
