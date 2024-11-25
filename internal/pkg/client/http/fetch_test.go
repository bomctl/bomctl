// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/http/fetch_test.go
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

package http_test

import (
	"context"
	"fmt"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/http"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/testutil"
)

type httpFetchSuite struct {
	suite.Suite
	*options.Options
	*db.Backend
	*http.Client
	*httptest.Server
	documentInfo []testutil.DocumentInfo
}

func (hfs *httpFetchSuite) SetupSuite() {
	fileServer := nethttp.FileServer(nethttp.Dir(testutil.GetTestdataDir()))
	hfs.Server = httptest.NewTLSServer(nethttp.StripPrefix("/testdata/", fileServer))
	hfs.Client = &http.Client{}

	hfs.Client.SetHTTPClient(hfs.Server.Client())
}

func (hfs *httpFetchSuite) SetupSubTest() {
	var err error

	hfs.Backend, err = testutil.NewTestBackend()
	hfs.Require().NoError(err, "failed database backend creation")

	hfs.documentInfo, err = testutil.AddTestDocuments(hfs.Backend)
	hfs.Require().NoError(err, "failed database backend setup")

	hfs.Options = options.New().WithContext(context.WithValue(context.Background(), db.BackendKey{}, hfs.Backend))
}

func (hfs *httpFetchSuite) TearDownSubTest() {
	hfs.Backend.CloseClient()
}

func (hfs *httpFetchSuite) TearDownSuite() {
	hfs.Server.Close()
	hfs.Backend.CloseClient()
}

func (hfs *httpFetchSuite) TestClient_Fetch() {
	for _, alias := range []string{"cdx", "spdx"} {
		want, err := os.ReadFile(filepath.Join(testutil.GetTestdataDir(), fmt.Sprintf("sbom.%s.json", alias)))
		hfs.Require().NoError(err)

		hfs.Run(alias, func() {
			fetchURL := fmt.Sprintf("%s/testdata/sbom.%s.json", hfs.Server.URL, alias)
			opts := &options.FetchOptions{Options: hfs.Options}

			got, err := hfs.Fetch(fetchURL, opts)
			hfs.Require().NoError(err)

			hfs.Len(got, len(want))

			document, err := hfs.GetDocumentByIDOrAlias(alias)
			hfs.Require().NoError(err)

			srcData, err := hfs.GetDocumentUniqueAnnotation(document.GetMetadata().GetId(), db.SourceDataAnnotation)
			hfs.Require().NoError(err)

			hfs.Len(srcData, len(want))
		})
	}
}

func TestHTTPFetchSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(httpFetchSuite))
}
