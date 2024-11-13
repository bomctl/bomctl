// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/outpututil/file_test.go
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

package outpututil_test

import (
	"bytes"
	"context"
	"path"
	"testing"

	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/outpututil"
	"github.com/bomctl/bomctl/internal/testutil"
)

type fileSuite struct {
	suite.Suite
	*db.Backend
	*options.Options

	documentInfo []testutil.DocumentInfo
}

func (fs *fileSuite) SetupSuite() {
	var err error

	fs.Backend, err = testutil.NewTestBackend()
	fs.Require().NoError(err, "failed database backend creation")

	fs.documentInfo, err = testutil.AddTestDocuments(fs.Backend)
	fs.Require().NoError(err, "failed database backend setup")

	fs.Options = options.New().WithContext(context.WithValue(context.Background(), db.BackendKey{}, fs.Backend))
}

func (fs *fileSuite) TearDownSuite() {
	fs.Backend.CloseClient()
}

func (fs *fileSuite) TestCheckIfModified() {
	tempbe, err := testutil.NewTestBackend()
	fs.Require().NoError(err, "failed database backend creation")

	tempdocs, err := testutil.AddTestDocuments(tempbe)
	fs.Require().NoError(err, "failed database backend setup")

	defer tempbe.CloseClient()

	for _, data := range []struct {
		prepare       func()
		name          string
		docNum        int
		expectedValue bool
	}{
		{
			name:          "Not Modified",
			docNum:        0,
			prepare:       func() {},
			expectedValue: false,
		},
		{
			name:   "Modified",
			docNum: 1,
			prepare: func() {
				docID := tempdocs[1].Document.GetMetadata().GetId()
				err := tempbe.ClearAnnotations(docID)
				fs.Require().NoError(err)
			},
			expectedValue: true,
		},
		{
			name:   "Empty SourceDataAnnotation annotation",
			docNum: 0,
			prepare: func() {
				docID := tempdocs[0].Document.GetMetadata().GetId()
				err := tempbe.SetUniqueAnnotation(docID, db.SourceDataAnnotation, "")
				fs.Require().NoError(err)
			},
			expectedValue: true,
		},
	} {
		fs.Run(data.name, func() {
			data.prepare()
			got, err := outpututil.CheckIfModified(tempdocs[data.docNum].Document, tempbe)
			fs.Require().NoError(err)

			fs.Require().Equal(data.expectedValue, got)
		})
	}
}

func (fs *fileSuite) TestMatchesOriginFormat() {
	for _, data := range []struct {
		name           string
		expectedFormat string
		format         formats.Format
		expectedValue  bool
	}{
		{
			name:           "spdx 2.3",
			expectedFormat: "text/spdx+json;version=2.3",
			format:         formats.SPDX23JSON,
			expectedValue:  true,
		},
		{
			name:           "spdx 2.3 text encoding",
			expectedFormat: "text/spdx+text;version=2.3",
			format:         formats.SPDX23JSON,
			expectedValue:  false,
		},
		{
			name:           "non-supported spdx",
			expectedFormat: "text/spdx+json;version=0.7",
			format:         formats.SPDX23JSON,
			expectedValue:  false,
		},
		{
			name:           "cyclonedx 1.5 json",
			expectedFormat: "application/vnd.cyclonedx+json;version=1.5",
			format:         formats.CDX15JSON,
			expectedValue:  true,
		},
		{
			name:           "cyclonedx 1.6 encoding mismatch",
			expectedFormat: "application/vnd.cyclonedx+xml;version=1.6",
			format:         formats.CDX16JSON,
			expectedValue:  false,
		},
		{
			name:           "cyclonedx 1.5 xml",
			expectedFormat: "application/vnd.cyclonedx+xml;version=1.5",
			format:         formats.Format("application/vnd.cyclonedx+xml;version=1.5"),
			expectedValue:  true,
		},
		{
			name:           "unknown encoding cyclonedx",
			expectedFormat: "application/vnd.cyclonedx+text;version=1.4",
			format:         formats.CDX16JSON,
			expectedValue:  false,
		},
	} {
		fs.Run(data.name, func() {
			document := sbom.Document{
				Metadata: &sbom.Metadata{
					SourceData: &sbom.SourceData{
						Format: data.expectedFormat,
					},
				},
			}

			got := outpututil.MatchesOriginFormat(&document, data.format)

			fs.Require().Equal(data.expectedValue, got)
		})
	}
}

func (fs *fileSuite) TestWriteOriginStream() {
	for _, docInfo := range fs.documentInfo {
		stream := testutil.FakeWriter{Buffer: &bytes.Buffer{}}
		err := outpututil.WriteOriginStream(docInfo.Document, fs.Backend, &stream)
		fs.Require().NoError(err)

		fs.Require().Equal(docInfo.Content, stream.Buffer.Bytes())
	}
}

func (fs *fileSuite) TestWriteStream() {
	for _, data := range []struct {
		name string
		fmt  formats.Format
	}{
		{
			name: "SPDX23JSON",
			fmt:  formats.SPDX23JSON,
		},
		{
			name: "CDX12JSON",
			fmt:  formats.CDX12JSON,
		},
		{
			name: "CDX13JSON",
			fmt:  formats.CDX13JSON,
		},
		{
			name: "CDX14JSON",
			fmt:  formats.CDX14JSON,
		},
		{
			name: "CDX15JSON",
			fmt:  formats.CDX15JSON,
		},
		{
			name: "CDX16JSON",
			fmt:  formats.CDX16JSON,
		},
	} {
		fs.Run(data.name, func() {
			for _, docInfo := range fs.documentInfo {
				stream := testutil.FakeWriter{Buffer: &bytes.Buffer{}}
				err := outpututil.WriteStream(docInfo.Document, data.fmt, fs.Options, &stream)
				fs.Require().NoError(err)
			}
		})
	}
}

func (fs *fileSuite) TestWriteFile() {
	for _, data := range []struct {
		name string
		fmt  formats.Format
	}{
		{
			name: "SPDX23JSON",
			fmt:  formats.SPDX23JSON,
		},
		{
			name: "CDX12JSON",
			fmt:  formats.CDX12JSON,
		},
		{
			name: "CDX13JSON",
			fmt:  formats.CDX13JSON,
		},
		{
			name: "CDX14JSON",
			fmt:  formats.CDX14JSON,
		},
		{
			name: "CDX15JSON",
			fmt:  formats.CDX15JSON,
		},
		{
			name: "CDX16JSON",
			fmt:  formats.CDX16JSON,
		},
	} {
		fs.Run(data.name, func() {
			for _, docInfo := range fs.documentInfo {
				tmpDir := fs.T().TempDir()
				p := path.Join(tmpDir, data.name+".json")
				err := outpututil.WriteFile(docInfo.Document, data.fmt, fs.Options, p)
				fs.Require().NoError(err)
			}
		})
	}
}

func TestFileSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(fileSuite))
}

// SetupTest runs before each test in the suite.
func (*fileSuite) SetupTest() {}

// SetupSubtTest runs before each subtest in the suite.
func (*fileSuite) SetupSubTest() {}

// BeforeTest is executed right before the test starts.
func (*fileSuite) BeforeTest(_, _ string) {}

// AfterTest is executed right after the test finishes.
func (*fileSuite) AfterTest(_, _ string) {}

// TearDownSubTest runs after each subtest in the suite have been run.
func (*fileSuite) TearDownSubTest() {}

// TearDownTest runs after each test in the suite.
func (*fileSuite) TearDownTest() {}
