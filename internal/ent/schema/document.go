// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: .ent/schema/document.go
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
	"entgo.io/ent/schema/edge"
)

type Document struct {
	ent.Schema
}

func (Document) Fields() []ent.Field { return nil }

func (Document) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("metadata", Metadata.Type).Unique().Immutable(),
		edge.To("node_list", NodeList.Type).Unique().Immutable(),
	}
}

func (Document) Annotations() []schema.Annotation { return nil }
