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
	"fmt"
	"path/filepath"

	"github.com/bom-squad/protobom/pkg/sbom"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type databaseORM struct {
	*gorm.DB
}

var db *databaseORM

// Create database and initialize schema.
func Create() (*databaseORM, error) {
	cacheDir := viper.GetString("cache_dir")
	dbFile := filepath.Join(cacheDir, "bomctl.db")

	dbConn, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error opening database file: %w", err)
	}

	db = &databaseORM{DB: dbConn}

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
