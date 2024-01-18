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
	"embed"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type databaseORM struct {
	*gorm.DB
}

var db *databaseORM

//go:embed scripts
var scripts embed.FS

// Create database and initialize schema.
func Create() (*databaseORM, error) {
	cacheDir := viper.GetString("cache_dir")
	dbFile := filepath.Join(cacheDir, "bomctl.db")

	dbConn, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error opening database file: %w", err)
	}

	db = &databaseORM{DB: dbConn}

	for _, scriptName := range []string{"types", "tables", "triggers", "views"} {
		script, err := scripts.ReadFile(filepath.Join("scripts", scriptName+".sql"))
		if err != nil {
			return nil, fmt.Errorf("error reading database %s creation script file: %w", scriptName, err)
		}

		db.Exec(string(script))
	}

	return db, nil
}
