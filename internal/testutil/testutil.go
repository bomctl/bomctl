// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/testutil/testutil.go
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

package testutil

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

type TestWriter struct {
	*bytes.Buffer
}

func (*TestWriter) Close() error {
	return nil
}

type DocumentInfo struct {
	Document *sbom.Document
	Content  []byte
}

var lock = sync.Mutex{} //nolint:gochecknoglobals

// AddTestDocuments preloads a Backend with SBOMs from the testdata directory.
// In addition, the stored Documents and corresponding bytes data are captured and returned.
func AddTestDocuments(backend *db.Backend) ([]DocumentInfo, error) {
	testdataDir := GetTestdataDir()

	sboms, err := os.ReadDir(testdataDir)
	if err != nil {
		return nil, fmt.Errorf("reading testdata directory: %w", err)
	}

	documentInfo := []DocumentInfo{}

	for idx := range sboms {
		data, err := os.ReadFile(filepath.Join(testdataDir, sboms[idx].Name()))
		if err != nil {
			return nil, fmt.Errorf("reading testdata file %s: %w", sboms[idx].Name(), err)
		}

		doc, err := backend.AddDocument(data)
		if err != nil {
			return nil, fmt.Errorf("storing document: %w", err)
		}

		name := strings.Split(sboms[idx].Name(), ".")[1]

		if err := backend.SetUniqueAnnotation(doc.GetMetadata().GetId(), db.AliasAnnotation, name); err != nil {
			return nil, fmt.Errorf("setting alias: %w", err)
		}

		tags := []string{"tag" + strconv.Itoa(idx+1), "tag" + strconv.Itoa(idx+2)} //nolint:mnd

		if err := backend.AddAnnotations(doc.GetMetadata().GetId(), db.TagAnnotation, tags...); err != nil {
			return nil, fmt.Errorf("adding tags: %w", err)
		}

		documentInfo = append(documentInfo, DocumentInfo{Document: doc, Content: data})
	}

	return documentInfo, nil
}

// NewTestBackend creates a Backend for testing with in-memory storage.
func NewTestBackend() (*db.Backend, error) {
	lock.Lock()
	defer lock.Unlock()

	backend, err := db.NewBackend(db.WithDatabaseFile(":memory:"))
	if err != nil {
		return nil, fmt.Errorf("database setup: %w", err)
	}

	return backend, nil
}

func GetTestdataDir() string {
	_, filename, _, _ := runtime.Caller(0) //nolint:dogsled

	return filepath.Join(filepath.Dir(filename), "testdata")
}
