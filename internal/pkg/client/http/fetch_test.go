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
}

func (hfs *httpFetchSuite) SetupSuite() {
	hfs.Server = httptest.NewTLSServer(nethttp.FileServer(nethttp.Dir(testutil.GetTestdataDir())))
	hfs.Client = &http.Client{}

	hfs.Client.SetHTTPClient(hfs.Server.Client())
}

func (hfs *httpFetchSuite) SetupSubTest() {
	var err error

	hfs.Backend, err = testutil.NewTestBackend()
	hfs.Require().NoError(err, "failed database backend creation")

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
		hfs.Run(alias, func() {
			fetchURL := fmt.Sprintf("%s/testdata/sbom.%s.json", hfs.Server.URL, alias)
			opts := &options.FetchOptions{Options: hfs.Options}

			data, err := hfs.Fetch(fetchURL, opts)
			hfs.Require().NoError(err)

			document, err := hfs.GetDocumentByIDOrAlias(alias)
			hfs.Require().NoError(err)

			annotations, err := hfs.GetDocumentAnnotations(document.GetMetadata().GetId(),
				db.SourceDataAnnotation,
				db.SourceFormatAnnotation,
				db.SourceURLAnnotation,
			)
			hfs.Require().NoError(err)

			for idx := range annotations {
				switch annotations[idx].Name {
				case db.SourceDataAnnotation:
					hfs.Len(annotations[idx].Value, len(data))
				case db.SourceFormatAnnotation:
					hfs.Contains(annotations[idx].Value, alias)
				case db.SourceURLAnnotation:
					hfs.Equal(annotations[idx].Value, fetchURL)
				}
			}
		})
	}
}

func TestHTTPFetchSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(httpFetchSuite))
}
