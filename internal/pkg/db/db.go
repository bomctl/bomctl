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

	protobom "github.com/bom-squad/protobom/pkg/sbom"
	"github.com/charmbracelet/log"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bomctl/bomctl/internal/pkg/utils"
)

// Enable SQLite foreign key support.
const dsnParams string = "?_pragma=foreign_keys(1)"

var (
	ctx    = context.Background()
	db     *gorm.DB
	logger *log.Logger
)

type ORMToPBConverter interface {
	ToPB(context.Context) (protobom.Document, error)
}

type PBToORMConverter interface {
	ToORM(context.Context) (protobom.DocumentORM, error)
}

// Create database and initialize schema.
func CreateSchema(dbFile string) (*gorm.DB, error) {
	logger = utils.NewLogger("")

	if db != nil {
		logger.Info("Database file already exists, will not recreate")
		return db, nil
	}

	logger.Info("Initializing database")
	logger.Debug("Connection string", "dbFile", dbFile, "dsnParams", dsnParams)

	var err error

	db, err = gorm.Open(sqlite.Open(dbFile + dsnParams))
	if err != nil {
		return nil, fmt.Errorf("error opening database file %s: %w", dbFile, err)
	}

	// Create database tables from model definitions.
	models := []interface{}{
		&protobom.DocumentORM{},
		&protobom.DocumentTypeORM{},
		&protobom.EdgeORM{},
		&protobom.ExternalReferenceORM{},
		&protobom.MetadataORM{},
		&protobom.NodeListORM{},
		&protobom.NodeORM{},
		&protobom.PersonORM{},
		&protobom.ToolORM{},
	}

	for _, model := range models {
		err := db.AutoMigrate(model)
		if err != nil {
			return nil, fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	return db, nil
}

// Insert protobom Document into `documents` table.
func AddDocument(document PBToORMConverter) error {
	documentORM, err := document.ToORM(ctx)
	if err != nil {
		return fmt.Errorf("failed to convert protobuf models to gorm: %w", err)
	}

	session := db.Session(&gorm.Session{Context: ctx})
	tx := session.Begin()
	tx.Clauses(clause.Insert{Modifier: "OR IGNORE"}).Create(&documentORM.Metadata)
	tx.Clauses(clause.Insert{Modifier: "OR IGNORE"}).Create(&documentORM.NodeList)
	tx.Clauses(clause.Insert{Modifier: "OR IGNORE"}).Create(&documentORM)
	tx.Commit()

	return nil
}

func GetDocumentByID(id uint32) *protobom.Document {
	documentORM := &protobom.DocumentORM{}

	db.Where(&protobom.DocumentORM{Id: id}).First(&documentORM)
	db.Where(&protobom.MetadataORM{DocumentId: &id}).First(&documentORM.Metadata)
	db.Where(&protobom.NodeListORM{DocumentId: &id}).First(&documentORM.NodeList)

	db.Where(&protobom.DocumentTypeORM{MetadataId: &documentORM.Metadata.Id}).Find(&documentORM.Metadata.DocumentTypes)
	db.Where(&protobom.PersonORM{MetadataId: &documentORM.Metadata.Id}).Find(&documentORM.Metadata.Authors)
	db.Where(&protobom.ToolORM{MetadataId: &documentORM.Metadata.Id}).Find(&documentORM.Metadata.Tools)

	db.Where(&protobom.NodeORM{NodeListId: &documentORM.NodeList.Id}).Find(&documentORM.NodeList.Nodes)
	db.Where(&protobom.EdgeORM{NodeListId: &documentORM.NodeList.Id}).Find(&documentORM.NodeList.Edges)

	document, err := documentORM.ToPB(ctx)
	if err != nil {
		return nil
	}

	return &document
}
