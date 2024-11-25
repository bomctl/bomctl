// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/merge/merger_test.go
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

package merge_test

import (
	"testing"
	"time"

	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bomctl/bomctl/internal/pkg/merge"
)

type mergerSuite struct {
	suite.Suite
}

type nodeOption struct {
	ID   string
	Name string
}

var defaultTime time.Time

func TestMergerSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(mergerSuite))
}

func (ms *mergerSuite) TestMergeWithWrongObject() {
	doc1 := &sbom.Document{}
	doc2 := &sbom.Document{}
	err := merge.NewMerger(doc1).MergeProtoMessage(doc2)

	ms.Error(err)
}

func (ms *mergerSuite) TestMergeMetadata() {
	for _, data := range []struct {
		base        *sbom.Metadata
		other       *sbom.Metadata
		expected    *sbom.Metadata
		expectedErr error
	}{
		{
			base:  createMetadata("1", nil, nil, nil),
			other: createMetadata("2", nil, nil, nil),
			expected: &sbom.Metadata{
				Name:    "metadata1",
				Version: "1",
				Comment: "metadata1",
				Date:    timestamppb.New(defaultTime),
				Id:      "metadata1",
			},
			expectedErr: nil,
		},
		{
			base:  &sbom.Metadata{},
			other: createMetadata("2", nil, nil, nil),
			expected: &sbom.Metadata{
				Name:    "metadata2",
				Version: "2",
				Comment: "metadata2",
				Date:    timestamppb.New(defaultTime),
				Id:      "metadata2",
			},
			expectedErr: nil,
		},
		{
			base: &sbom.Metadata{
				Name:    "metadata1",
				Version: "1",
				Comment: "",
				Date:    timestamppb.New(defaultTime),
				Id:      "metadata1",
			},
			other: createMetadata("2", nil, nil, nil),
			expected: &sbom.Metadata{
				Name:    "metadata1",
				Version: "1",
				Comment: "metadata2",
				Date:    timestamppb.New(defaultTime),
				Id:      "metadata1",
			},
			expectedErr: nil,
		},
	} {
		err := merge.NewMerger(data.base).MergeProtoMessage(data.other)

		ms.Equal(data.expectedErr, err)
		ms.Equal(data.expected, data.base)
	}
}

func (ms *mergerSuite) TestMergeMetadataSlices() {
	for _, data := range []struct {
		base        *sbom.Metadata
		other       *sbom.Metadata
		expected    *sbom.Metadata
		expectedErr error
	}{
		{
			base: createMetadata("1",
				[]*sbom.Tool{
					{
						Name:    "test1",
						Version: "test1",
						Vendor:  "test1",
					},
					{
						Name:    "test2",
						Version: "test2",
						Vendor:  "test2",
					},
				},
				[]*sbom.Person{
					createPerson("1"),
					createPerson("2"),
				},
				[]*sbom.DocumentType{
					createDocumentType("test1", "test1"),
					createDocumentType("test2", "test2"),
				},
			),
			other: createMetadata("2", nil, nil, nil),
			expected: &sbom.Metadata{
				Name:    "metadata1",
				Version: "1",
				Comment: "metadata1",
				Date:    timestamppb.New(defaultTime),
				Id:      "metadata1",
				Tools: []*sbom.Tool{
					{
						Name:    "test1",
						Version: "test1",
						Vendor:  "test1",
					},
					{
						Name:    "test2",
						Version: "test2",
						Vendor:  "test2",
					},
				},
				Authors: []*sbom.Person{
					createPerson("1"),
					createPerson("2"),
				},
				DocumentTypes: []*sbom.DocumentType{
					createDocumentType("test1", "test1"),
					createDocumentType("test2", "test2"),
				},
			},
			expectedErr: nil,
		},
		{
			base: createMetadata("1",
				[]*sbom.Tool{
					{
						Name:    "test1",
						Version: "test1",
						Vendor:  "",
					},
					{
						Name:    "test1",
						Version: "test1",
						Vendor:  "test2",
					},
				},
				[]*sbom.Person{
					createPerson("1"),
					{
						Name:  "person2",
						Email: "person2@test.com",
						Phone: "",
						Url:   "",
					},
					{
						Name:  "person2",
						Email: "person2@test.com",
						Phone: "0123456789",
						Url:   "",
					},
					{
						Name:  "person2",
						Email: "person2@test.com",
						Phone: "",
						Url:   "http://person2.com",
					},
				},
				[]*sbom.DocumentType{
					createDocumentType("test1", ""),
					createDocumentType("test1", "test2"),
				},
			),
			other: createMetadata("2", nil, nil, nil),
			expected: &sbom.Metadata{
				Name:    "metadata1",
				Version: "1",
				Comment: "metadata1",
				Date:    timestamppb.New(defaultTime),
				Id:      "metadata1",
				Tools: []*sbom.Tool{
					{
						Name:    "test1",
						Version: "test1",
						Vendor:  "test2",
					},
				},
				Authors: []*sbom.Person{
					createPerson("1"),
					createPerson("2"),
				},
				DocumentTypes: []*sbom.DocumentType{
					createDocumentType("test1", "test2"),
				},
			},
			expectedErr: nil,
		},
	} {
		err := merge.NewMerger(data.base).MergeProtoMessage(data.other)

		ms.Equal(data.expectedErr, err)
		ms.Equal(data.expected, data.base)
	}
}

func (ms *mergerSuite) TestMergeNode() {
	for _, data := range []struct {
		base        *sbom.Node
		other       *sbom.Node
		expected    *sbom.Node
		expectedErr error
	}{
		{
			base: &sbom.Node{
				Id:   "node1",
				Name: "node1",
			},
			other: &sbom.Node{
				Id:   "node2",
				Name: "node2",
			},
			expected: &sbom.Node{
				Id:   "node1",
				Name: "node1",
			},
			expectedErr: nil,
		},
		{
			base: &sbom.Node{},
			other: &sbom.Node{
				Name: "node2",
			},
			expected: &sbom.Node{
				Name: "node2",
			},
			expectedErr: nil,
		},
	} {
		err := merge.NewMerger(data.base).MergeProtoMessage(data.other)

		ms.Equal(data.expectedErr, err)
		ms.Equal(data.expected, data.base)
	}
}

func (ms *mergerSuite) TestMergeNodeList() {
	for _, data := range []struct {
		base        *sbom.NodeList
		other       *sbom.NodeList
		expected    *sbom.NodeList
		expectedErr error
	}{
		{
			base: createNodeList(
				[]nodeOption{
					{
						ID:   "test1",
						Name: "test1",
					},
				},
			),
			other: createNodeList(
				[]nodeOption{
					{
						ID:   "test2",
						Name: "test2",
					},
				},
			),
			expected: createNodeList(
				[]nodeOption{
					{
						ID:   "test1",
						Name: "test1",
					},
					{
						ID:   "test2",
						Name: "test2",
					},
				},
			),
			expectedErr: nil,
		},
		{
			base: createNodeList(
				[]nodeOption{
					{
						ID:   "test1",
						Name: "",
					},
				},
			),
			other: createNodeList(
				[]nodeOption{
					{
						ID:   "test1",
						Name: "test2",
					},
				},
			),
			expected: createNodeList(
				[]nodeOption{
					{
						ID:   "test1",
						Name: "test2",
					},
				},
			),
			expectedErr: nil,
		},
	} {
		err := merge.NewMerger(data.base).MergeProtoMessage(data.other)

		ms.Equal(data.expectedErr, err)
		ms.Equal(data.expected, data.base)
	}
}

func (ms *mergerSuite) TestMergePerson() {
	for _, data := range []struct {
		base        *sbom.Person
		other       *sbom.Person
		expected    *sbom.Person
		expectedErr error
	}{
		{
			base:  createPerson("1"),
			other: createPerson("2"),
			expected: &sbom.Person{
				Name:     "person1",
				Email:    "person1@test.com",
				Url:      "http://person1.com",
				Phone:    "0123456789",
				Contacts: nil,
			},
			expectedErr: nil,
		},
		{
			base:  &sbom.Person{},
			other: createPerson("2"),
			expected: &sbom.Person{
				Name:     "person2",
				Email:    "person2@test.com",
				Url:      "http://person2.com",
				Phone:    "0123456789",
				Contacts: nil,
			},
			expectedErr: nil,
		},
		{
			base: &sbom.Person{
				Name:  "person1",
				Email: "person1@test.com",
				Url:   "",
				Phone: "0123456789",
			},
			other: createPerson("2"),
			expected: &sbom.Person{
				Name:     "person1",
				Email:    "person1@test.com",
				Url:      "http://person2.com",
				Phone:    "0123456789",
				Contacts: nil,
			},
			expectedErr: nil,
		},
	} {
		err := merge.NewMerger(data.base).MergeProtoMessage(data.other)

		ms.Equal(data.expectedErr, err)
		ms.Equal(data.expected, data.base)
	}
}

func (ms *mergerSuite) TestMergeTool() {
	for _, data := range []struct {
		base        *sbom.Tool
		other       *sbom.Tool
		expected    *sbom.Tool
		expectedErr error
	}{
		{
			base: &sbom.Tool{
				Name:    "test1",
				Version: "test1",
				Vendor:  "test1",
			},
			other: &sbom.Tool{
				Name:    "test2",
				Version: "test2",
				Vendor:  "test2",
			},
			expected: &sbom.Tool{
				Name:    "test1",
				Version: "test1",
				Vendor:  "test1",
			},
			expectedErr: nil,
		},
		{
			base: &sbom.Tool{},
			other: &sbom.Tool{
				Name:    "test2",
				Version: "test2",
				Vendor:  "test2",
			},
			expected: &sbom.Tool{
				Name:    "test2",
				Version: "test2",
				Vendor:  "test2",
			},
			expectedErr: nil,
		},
		{
			base: &sbom.Tool{
				Name:    "test1",
				Version: "test1",
				Vendor:  "",
			},
			other: &sbom.Tool{
				Name:    "test2",
				Version: "test2",
				Vendor:  "test2",
			},
			expected: &sbom.Tool{
				Name:    "test1",
				Version: "test1",
				Vendor:  "test2",
			},
			expectedErr: nil,
		},
	} {
		err := merge.NewMerger(data.base).MergeProtoMessage(data.other)

		ms.Equal(data.expectedErr, err)
		ms.Equal(data.expected, data.base)
	}
}

func (ms *mergerSuite) TestMergeDcoumentType() {
	for _, data := range []struct {
		base        *sbom.DocumentType
		other       *sbom.DocumentType
		expected    *sbom.DocumentType
		expectedErr error
	}{
		{
			base:        createDocumentType("test1", "test1"),
			other:       createDocumentType("test2", "test1"),
			expected:    createDocumentType("test1", "test1"),
			expectedErr: nil,
		},
		{
			base:        &sbom.DocumentType{},
			other:       createDocumentType("test2", "test2"),
			expected:    createDocumentType("test2", "test2"),
			expectedErr: nil,
		},
		{
			base:        createDocumentType("test1", ""),
			other:       createDocumentType("test2", "test2"),
			expected:    createDocumentType("test1", "test2"),
			expectedErr: nil,
		},
	} {
		err := merge.NewMerger(data.base).MergeProtoMessage(data.other)

		ms.Equal(data.expectedErr, err)
		ms.Equal(data.expected, data.base)
	}
}

func createNodeList(opts []nodeOption) *sbom.NodeList {
	nodeList := sbom.NewNodeList()

	for _, opt := range opts {
		node := sbom.NewNode()
		node.Id = opt.ID
		node.Name = opt.Name
		nodeList.AddNode(node)
	}

	return nodeList
}

func createMetadata(val string, tools []*sbom.Tool, authors []*sbom.Person,
	docTypes []*sbom.DocumentType,
) *sbom.Metadata {
	return &sbom.Metadata{
		Name:          "metadata" + val,
		Id:            "metadata" + val,
		Date:          timestamppb.New(defaultTime),
		Version:       val,
		Comment:       "metadata" + val,
		Tools:         tools,
		Authors:       authors,
		DocumentTypes: docTypes,
	}
}

func createPerson(val string) *sbom.Person {
	return &sbom.Person{
		Name:  "person" + val,
		Email: "person" + val + "@test.com",
		Phone: "0123456789",
		Url:   "http://person" + val + ".com",
	}
}

func createDocumentType(name, description string) *sbom.DocumentType {
	return &sbom.DocumentType{
		Name:        &name,
		Type:        sbom.DocumentType_OTHER.Enum(),
		Description: &description,
	}
}
