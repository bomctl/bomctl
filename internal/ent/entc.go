//go:build ignore
// +build ignore

// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/ent/entc.go
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
package main

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	"github.com/bomctl/bomctl/internal/pkg/utils"
)

func main() {
	logger := utils.NewLogger("ent")

	if err := entc.Generate("./schema", &gen.Config{
		Features:  []gen.Feature{gen.FeatureUpsert},
		Templates: []*gen.Template{gen.MustParse(gen.NewTemplate("header").ParseFiles("template/header.tmpl"))},
	}); err != nil {
		logger.Fatalf("running ent codegen: %v", err)
	}
}
