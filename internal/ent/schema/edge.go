// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: .ent/schema/edge.go
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
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Edge struct {
	ent.Schema
}

func (Edge) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").Values(
			"UNKNOWN",
			"amends",
			"ancestor",
			"buildDependency",
			"buildTool",
			"contains",
			"contained_by",
			"copy",
			"dataFile",
			"dependencyManifest",
			"dependsOn",
			"dependencyOf",
			"descendant",
			"describes",
			"describedBy",
			"devDependency",
			"devTool",
			"distributionArtifact",
			"documentation",
			"dynamicLink",
			"example",
			"expandedFromArchive",
			"fileAdded",
			"fileDeleted",
			"fileModified",
			"generates",
			"generatedFrom",
			"metafile",
			"optionalComponent",
			"optionalDependency",
			"other",
			"packages",
			"patch",
			"prerequisite",
			"prerequisiteFor",
			"providedDependency",
			"requirementFor",
			"runtimeDependency",
			"specificationFor",
			"staticLink",
			"test",
			"testCase",
			"testDependency",
			"testTool",
			"variant",
		),
		field.String("from"),
		field.String("to"),
	}
}

func (Edge) Edges() []ent.Edge { return nil }

func (Edge) Annotations() []schema.Annotation { return nil }
