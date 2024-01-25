// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/db/db.go
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
package db

import (
	"context"
	"fmt"

	"github.com/bom-squad/protobom/pkg/sbom"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Enable SQLite foreign key support.
const dsnParams string = "?_pragma=foreign_keys(1)"

var (
	ctx = context.Background()
	db  *gorm.DB
)

// Create database and initialize schema.
func CreateSchema(dbFile string) (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	var err error

	db, err = gorm.Open(sqlite.Open(dbFile+dsnParams), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening database file %s: %w", dbFile, err)
	}

	// Create database tables from model definitions.
	models := []interface{}{
		&sbom.DocumentORM{},
		&sbom.DocumentTypeORM{},
		&sbom.EdgeORM{},
		&sbom.ExternalReferenceORM{},
		&sbom.MetadataORM{},
		&sbom.NodeListORM{},
		&sbom.NodeORM{},
		&sbom.PersonORM{},
		&sbom.ToolORM{},
	}

	for _, model := range models {
		err := db.AutoMigrate(model)
		if err != nil {
			return nil, fmt.Errorf("%T: %w", model, err)
		}
	}

	return db, nil
}

// Insert protobom Document into `documents` table.
func AddDocument(document *sbom.Document) error {
	documentORM, err := document.ToORM(ctx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	documentORM.Id = document.Metadata.Id

	db.Clauses(clause.OnConflict{DoNothing: true}).Create(documentORM)

	return nil
}

func GetDocumentByID(id string) *sbom.Document {
	documentORM := &sbom.DocumentORM{}

	db.Where(&sbom.DocumentORM{Id: id}).First(&documentORM)
	db.Where(&sbom.MetadataORM{DocumentId: &id}).First(&documentORM.Metadata)
	db.Where(&sbom.NodeListORM{DocumentId: &id}).First(&documentORM.NodeList)

	db.Where(&sbom.DocumentTypeORM{MetadataId: &documentORM.Metadata.Id}).Find(&documentORM.Metadata.DocumentTypes)
	db.Where(&sbom.PersonORM{MetadataId: &documentORM.Metadata.Id}).Find(&documentORM.Metadata.Authors)
	db.Where(&sbom.ToolORM{MetadataId: &documentORM.Metadata.Id}).Find(&documentORM.Metadata.Tools)

	db.Where(&sbom.NodeORM{NodeListId: &documentORM.NodeList.Id}).Find(&documentORM.NodeList.Nodes)
	db.Where(&sbom.EdgeORM{NodeListId: &documentORM.NodeList.Id}).Find(&documentORM.NodeList.Edges)

	document, err := documentORM.ToPB(ctx)
	if err != nil {
		return nil
	}

	return &document
}
