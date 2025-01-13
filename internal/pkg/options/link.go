// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/options/link.go
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

package options

import "fmt"

const (
	LinkTargetTypeNode LinkTargetType = iota
	LinkTargetTypeDocument
)

type (
	LinkTargetType uint8

	LinkTarget struct {
		ID, Alias string
		Type      LinkTargetType
	}

	Link struct {
		From LinkTarget
		To   []LinkTarget
	}
)

func (lt *LinkTarget) String() string {
	str := lt.ID
	if lt.Type == LinkTargetTypeDocument && lt.Alias != "" && lt.Alias != str {
		str += fmt.Sprintf(" (%s)", lt.Alias)
	}

	return str
}

func (lt LinkTargetType) String() string {
	switch lt {
	case LinkTargetTypeDocument:
		return "document"
	case LinkTargetTypeNode:
		return "node"
	}

	return ""
}
