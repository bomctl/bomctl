// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/fetch_test.go
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

package git_test

import (
	"fmt"
	neturl "net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/testutil"
)

type gitFetchSuite struct {
	gitClientSuite
}

func (gfs *gitFetchSuite) TestClient_Fetch() {
	for _, alias := range []string{"cdx", "spdx"} {
		want, err := os.ReadFile(filepath.Join(testutil.GetTestdataDir(), fmt.Sprintf("sbom.%s.json", alias)))
		gfs.Require().NoError(err)

		gfs.Run(alias, func() {
			opts := &options.FetchOptions{Options: gfs.Options}

			fetchString := fmt.Sprintf("%s/test/repo.git?ref=main#path/to/sbom.%s.json", gfs.Server.URL, alias)
			fetchURL, err := neturl.Parse(fetchString)
			gfs.Require().NoError(err)

			gfs.Require().NoError(gfs.PrepareFetch(fetchURL, netutil.NewBasicAuth("", ""), opts.Options))

			got, err := gfs.Fetch(fetchString, opts)
			gfs.Require().NoError(err)

			gfs.Len(got, len(want))

			document, err := gfs.GetDocumentByIDOrAlias(alias)
			gfs.Require().NoError(err)

			srcData, err := gfs.GetDocumentUniqueAnnotation(document.GetMetadata().GetId(), db.SourceDataAnnotation)
			gfs.Require().NoError(err)

			gfs.Len(srcData, len(want))
		})
	}
}

func TestGitFetchSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(gitFetchSuite))
}
