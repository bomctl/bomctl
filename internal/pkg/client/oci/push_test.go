// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/push_test.go
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

package oci_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/oci"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

type ociPushSuite struct {
	suite.Suite
	tmpDir string
	*options.Options
	*db.Backend
	*oci.Client
	*httptest.Server
	docs []*sbom.Document
}

func (ops *ociPushSuite) setupOCIRepository() {
	ops.T().Helper()

	// Create server root and test repository directories.
	serverRoot := filepath.Join(ops.tmpDir, "oci-test-server")
	repoDir := filepath.Join(serverRoot, "test", "repo")
	ops.Require().NoError(os.MkdirAll(repoDir, os.ModeDir|os.ModePerm))

	ops.Server = httptest.NewServer(http.FileServer(http.Dir(serverRoot)))

	serverURL := ops.Client.Parse(fmt.Sprintf("%s/test/repo:0.0.1", strings.TrimPrefix(ops.Server.URL, "http://")))
	ops.Require().NotNil(serverURL)

	serverURL.Path = "/test/repo"

	err := ops.Client.CreateRepository(serverURL, nil)
	ops.Require().NoError(err)
}

func (ops *ociPushSuite) SetupSuite() {
	var err error

	ops.tmpDir, err = os.MkdirTemp("", "oci-push-test")
	ops.Require().NoErrorf(err, "Failed to create temporary directory: %v", err)

	ops.Backend, err = db.NewBackend(db.WithDatabaseFile(filepath.Join(ops.tmpDir, db.DatabaseFile)))
	ops.Require().NoError(err)

	ops.Client = &oci.Client{}

	ops.setupOCIRepository()

	testdataDir := filepath.Join("..", "..", "db", "testdata")

	sboms, err := os.ReadDir(testdataDir)
	if err != nil {
		ops.T().Fatalf("%v", err)
	}

	for idx := range sboms {
		sbomData, err := os.ReadFile(filepath.Join(testdataDir, sboms[idx].Name()))
		if err != nil {
			ops.T().Fatalf("%v", err)
		}

		doc, err := ops.Backend.AddDocument(sbomData)
		if err != nil {
			ops.FailNow("failed storing document", "err", err)
		}

		ops.docs = append(ops.docs, doc)
	}

	ops.Options = options.New().
		WithCacheDir(ops.tmpDir).
		WithContext(context.WithValue(context.Background(), db.BackendKey{}, ops.Backend))
}

func (ops *ociPushSuite) BeforeTest(_suiteName, _testName string) {
	ops.Require().NoError(
		ops.Client.PreparePush(
			fmt.Sprintf("%s/test/repo:0.0.1", strings.TrimPrefix(ops.Server.URL, "http://")),
			&options.PushOptions{Options: ops.Options},
		),
	)
}

func (ops *ociPushSuite) TearDownSuite() {
	ops.Server.Close()
	ops.Backend.CloseClient()

	if err := os.RemoveAll(ops.tmpDir); err != nil {
		ops.T().Fatalf("Error removing temp directory %s", db.DatabaseFile)
	}
}

func (ops *ociPushSuite) TestClient_AddFile() {
	// Test adding all SBOM files to artifact archive.
	for _, document := range ops.docs {
		ops.Require().NoError(ops.Client.AddFile(
			fmt.Sprintf("%s/test/repo:0.0.1", strings.TrimPrefix(ops.Server.URL, "http://")),
			document.GetMetadata().GetId(),
			&options.PushOptions{Options: ops.Options, Format: formats.SPDX23JSON},
		))
	}

	ops.Require().Len(ops.Descriptors(), len(ops.docs))

	for _, descriptor := range ops.Descriptors() {
		exists, err := ops.Client.Store().Exists(ops.Ctx(), descriptor)
		ops.Require().NoError(err)
		ops.True(exists)
	}
}

func TestOCIPushSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ociPushSuite))
}
