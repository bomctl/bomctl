// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/merge/merger.go
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
package merge

import (
	"errors"
	"fmt"

	"github.com/protobom/protobom/pkg/sbom"
	"google.golang.org/protobuf/proto"
)

type (
	// Merger is the interface for merging proto.Message types.
	Merger[T proto.Message] interface {
		MergeProtoMessage(T) error
	}

	// MergerBase is used to implement the generic Merger interface.
	MergerBase[T proto.Message] struct {
		Base T
	}
)

var (
	errMergeFailure = errors.New("failed to merge object")
	errCastFailure  = errors.New("failed to cast object")
)

func (m *MergerBase[T]) MergeProtoMessage(t T) (err error) {
	switch base := any(m.Base).(type) {
	case *sbom.Metadata:
		return mergeMetadata(base, t)
	case *sbom.Node:
		return mergeNode(base, t)
	case *sbom.NodeList:
		return mergeNodeList(base, t)
	case *sbom.Person:
		return mergePerson(base, t)
	case *sbom.Tool:
		return mergeTool(base, t)
	case *sbom.DocumentType:
		return mergeDocumentType(base, t)
	default:
		return fmt.Errorf("%w: %T", errMergeFailure, t)
	}
}

func mergeMetadata(base *sbom.Metadata, t proto.Message) error {
	other, ok := t.(*sbom.Metadata)
	if !ok {
		return fmt.Errorf("%w: %T", errCastFailure, other)
	}

	base.Comment = mergeStrings(base.Comment, other.Comment)
	base.Id = mergeStrings(base.Id, other.Id)
	base.Name = mergeStrings(base.Name, other.Name)
	base.Version = mergeStrings(base.Version, other.Version)

	if base.Date == nil && other.Date != nil {
		base.Date = other.Date
	}

	err := mergeMetadataSlices(base, other)
	if err != nil {
		return err
	}

	return nil
}

func mergeMetadataSlices(base, other *sbom.Metadata) error {
	var err error

	base.Tools, err = mergeTools(base.Tools, other.Tools)
	if err != nil {
		return err
	}

	base.Authors, err = mergePersons(base.Authors, other.Authors)
	if err != nil {
		return err
	}

	base.DocumentTypes, err = mergeDocumentTypes(base.DocumentTypes, other.DocumentTypes)

	return err
}

func mergeTools(base, other []*sbom.Tool) ([]*sbom.Tool, error) {
	var mergedList []*sbom.Tool
	mergedList = append(mergedList, base...)
	mergedList = append(mergedList, other...)

	return dedupeTools(mergedList)
}

func dedupeTools(tools []*sbom.Tool) ([]*sbom.Tool, error) {
	var dedupedList []*sbom.Tool

	toolMap := make(map[string]*sbom.Tool)

	for _, tool := range tools {
		key := fmt.Sprintf("%s-%s", tool.Name, tool.Version)
		if _, exists := toolMap[key]; !exists {
			toolMap[key] = tool
			dedupedList = append(dedupedList, tool)
		} else {
			existingTool := toolMap[key]

			err := mergeTool(existingTool, tool)
			if err != nil {
				return nil, err
			}
		}
	}

	return dedupedList, nil
}

func mergeNode(base *sbom.Node, t proto.Message) error {
	other, ok := any(t).(*sbom.Node)
	if !ok {
		return fmt.Errorf("%w: %T", errCastFailure, t)
	}

	base.Augment(other)

	return nil
}

func mergeNodeList(base *sbom.NodeList, t proto.Message) error {
	other, ok := any(t).(*sbom.NodeList)
	if !ok {
		return fmt.Errorf("%w: %T", errCastFailure, t)
	}

	mergedNodeList := base.Union(other)

	base.Nodes = mergedNodeList.Nodes
	base.Edges = mergedNodeList.Edges
	base.RootElements = mergedNodeList.RootElements

	return nil
}

func mergePerson(base *sbom.Person, t proto.Message) error {
	other, ok := any(t).(*sbom.Person)
	if !ok {
		return fmt.Errorf("%w: %T", errCastFailure, t)
	}

	base.Email = mergeStrings(base.Email, other.Email)
	base.Name = mergeStrings(base.Name, other.Name)
	base.Phone = mergeStrings(base.Phone, other.Phone)
	base.Url = mergeStrings(base.Url, other.Url)

	var err error
	base.Contacts, err = mergePersons(base.Contacts, other.Contacts)

	return err
}

func mergePersons(base, other []*sbom.Person) ([]*sbom.Person, error) {
	var mergedList []*sbom.Person
	mergedList = append(mergedList, base...)
	mergedList = append(mergedList, other...)

	return dedupePersons(mergedList)
}

func dedupePersons(persons []*sbom.Person) ([]*sbom.Person, error) {
	var dedupedList []*sbom.Person

	personMap := make(map[string]*sbom.Person)

	for _, person := range persons {
		email := person.Email
		if _, exists := personMap[email]; !exists {
			personMap[email] = person
			dedupedList = append(dedupedList, person)
		} else {
			existingPerson := personMap[email]

			err := mergePerson(existingPerson, person)
			if err != nil {
				return nil, err
			}
		}
	}

	return dedupedList, nil
}

func mergeTool(base *sbom.Tool, t proto.Message) error {
	other, ok := any(t).(*sbom.Tool)
	if !ok {
		return fmt.Errorf("%w: %T", errCastFailure, t)
	}

	base.Name = mergeStrings(base.Name, other.Name)
	base.Vendor = mergeStrings(base.Vendor, other.Vendor)
	base.Version = mergeStrings(base.Version, other.Version)

	return nil
}

func mergeDocumentType(base *sbom.DocumentType, t proto.Message) error {
	other, ok := any(t).(*sbom.DocumentType)
	if !ok {
		return fmt.Errorf("%w: %T", errCastFailure, t)
	}

	if (base.Name == nil || *base.Name == "") && other.Name != nil {
		base.Name = other.Name
	}

	if base.Type == nil && other.Type != nil {
		base.Type = other.Type
	}

	if (base.Description == nil || *base.Description == "") && other.Description != nil {
		base.Description = other.Description
	}

	return nil
}

func mergeDocumentTypes(base, other []*sbom.DocumentType) ([]*sbom.DocumentType, error) {
	mergedList := []*sbom.DocumentType{}
	mergedList = append(mergedList, base...)
	mergedList = append(mergedList, other...)

	return dedupeDocumentTypes(mergedList)
}

func dedupeDocumentTypes(documentTypes []*sbom.DocumentType) ([]*sbom.DocumentType, error) {
	var dedupedList []*sbom.DocumentType

	documentTypeMap := make(map[string]*sbom.DocumentType)

	for _, documentType := range documentTypes {
		key := *documentType.Name
		if _, exists := documentTypeMap[key]; !exists {
			documentTypeMap[key] = documentType
			dedupedList = append(dedupedList, documentType)
		} else {
			existingDocumentType := documentTypeMap[key]

			err := mergeDocumentType(existingDocumentType, documentType)
			if err != nil {
				return nil, err
			}
		}
	}

	return dedupedList, nil
}

func mergeStrings(base, other string) string {
	if base == "" && other != "" {
		base = other
	}

	return base
}

func NewMerger[T proto.Message](data T) *MergerBase[T] {
	return &MergerBase[T]{Base: data}
}

// Enforce implementation of interface at compile time.
var (
	_ Merger[*sbom.Node]         = (*MergerBase[*sbom.Node])(nil)
	_ Merger[*sbom.NodeList]     = (*MergerBase[*sbom.NodeList])(nil)
	_ Merger[*sbom.Metadata]     = (*MergerBase[*sbom.Metadata])(nil)
	_ Merger[*sbom.Person]       = (*MergerBase[*sbom.Person])(nil)
	_ Merger[*sbom.Tool]         = (*MergerBase[*sbom.Tool])(nil)
	_ Merger[*sbom.DocumentType] = (*MergerBase[*sbom.DocumentType])(nil)
)
