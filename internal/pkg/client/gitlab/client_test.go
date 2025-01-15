// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/gitlab/client_test.go
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

package gitlab_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/bomctl/bomctl/internal/pkg/client/gitlab"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/outpututil"
	"github.com/bomctl/bomctl/internal/testutil"
)

type (
	gitLabClientSuite struct {
		suite.Suite
		tmpDir string
		*options.Options
		*db.Backend
		documents    []*sbom.Document
		documentInfo []testutil.DocumentInfo
	}

	mockProjectProvider struct {
		mock.Mock
	}

	mockBranchProvider struct {
		mock.Mock
	}

	mockCommitProvider struct {
		mock.Mock
	}

	mockDependencyListExporter struct {
		mock.Mock
	}

	mockGenericPackagePublisher struct {
		mock.Mock
	}
)

//revive:disable:unchecked-type-assertion,import-shadowing

func (mpp *mockGenericPackagePublisher) PublishPackageFile(
	pid any,
	packageName, packageVersion, fileName string,
	content io.Reader,
	opt *gogitlab.PublishPackageFileOptions,
	options ...gogitlab.RequestOptionFunc, //nolint:gocritic
) (*gogitlab.GenericPackagesFile, *gogitlab.Response, error) {
	args := mpp.Called(pid, packageName, packageVersion, fileName, content, opt, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.GenericPackagesFile), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mpp *mockProjectProvider) GetProject(
	pid any,
	opt *gogitlab.GetProjectOptions,
	options ...gogitlab.RequestOptionFunc, //nolint:gocritic
) (*gogitlab.Project, *gogitlab.Response, error) {
	args := mpp.Called(pid, opt, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.Project), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mbp *mockBranchProvider) GetBranch(
	pid any,
	branch string,
	options ...gogitlab.RequestOptionFunc, //nolint:gocritic
) (*gogitlab.Branch, *gogitlab.Response, error) {
	args := mbp.Called(pid, branch, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.Branch), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mcp *mockCommitProvider) GetCommit(
	pid any,
	sha string,
	opt *gogitlab.GetCommitOptions,
	options ...gogitlab.RequestOptionFunc, //nolint:gocritic
) (*gogitlab.Commit, *gogitlab.Response, error) {
	args := mcp.Called(pid, sha, opt, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.Commit), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mdle *mockDependencyListExporter) CreateDependencyListExport(
	pipelineID int,
	opt *gogitlab.CreateDependencyListExportOptions,
	options ...gogitlab.RequestOptionFunc, //nolint:gocritic
) (*gogitlab.DependencyListExport, *gogitlab.Response, error) {
	args := mdle.Called(pipelineID, opt, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.DependencyListExport), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mdle *mockDependencyListExporter) GetDependencyListExport(
	id int,
	options ...gogitlab.RequestOptionFunc, //nolint:gocritic
) (*gogitlab.DependencyListExport, *gogitlab.Response, error) {
	args := mdle.Called(id, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.DependencyListExport), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mdle *mockDependencyListExporter) DownloadDependencyListExport(
	id int,
	options ...gogitlab.RequestOptionFunc, //nolint:gocritic
) (io.Reader, *gogitlab.Response, error) {
	args := mdle.Called(id, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(io.Reader), args.Get(1).(*gogitlab.Response), args.Error(2)
}

//revive:enable:unchecked-type-assertion,import-shadowing

func (glcs *gitLabClientSuite) SetupTest() {
	var err error

	glcs.tmpDir, err = os.MkdirTemp("", "gitlab-push-test")
	glcs.Require().NoError(err, "Failed to create temporary directory")

	glcs.Backend, err = testutil.NewTestBackend()
	glcs.Require().NoError(err, "failed database backend creation")

	glcs.documentInfo, err = testutil.AddTestDocuments(glcs.Backend)
	glcs.Require().NoError(err, "failed database backend setup")

	for _, docInfo := range glcs.documentInfo {
		glcs.documents = append(glcs.documents, docInfo.Document)
	}

	glcs.Options = options.New().
		WithCacheDir(glcs.tmpDir).
		WithContext(context.WithValue(context.Background(), db.BackendKey{}, glcs.Backend))
}

func (glcs *gitLabClientSuite) TearDownTest() {
	glcs.Backend.CloseClient()
	glcs.documents = nil
	glcs.documentInfo = nil

	if err := os.RemoveAll(glcs.tmpDir); err != nil {
		glcs.T().Fatalf("Error removing temp directory %s", glcs.tmpDir)
	}
}

var successGitLabResponse = &gogitlab.Response{
	Response: &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.0",
	},
}

func (glcs *gitLabClientSuite) TestClient_Fetch() {
	dummyProjectID := 1234
	dummyProjectName := "DUMMY_PROJECT"
	dummyBranchName := "DUMMY_BRANCH"
	dummyCommitSHA := "ABCD"
	dummyPipelineID := 2345
	dummyExportID := 3456

	dummyFetchURL := fmt.Sprintf("https://TEST_GITLAB.test/%s@%s", dummyProjectName, dummyBranchName)

	expectedCreateDependencyListExport := &gogitlab.DependencyListExport{
		ID:          dummyExportID,
		HasFinished: false,
		Self:        "TEST",
		Download:    "TEST/Download",
	}

	expectedGetDependencyListExport := &gogitlab.DependencyListExport{
		ID:          dummyExportID,
		HasFinished: true,
		Self:        "TEST",
		Download:    "TEST/Download",
	}

	expectedSbomData := []byte("DUMMY SBOM DATA")

	mockedProjectProvider := &mockProjectProvider{}
	mockedBranchProvider := &mockBranchProvider{}
	mockedCommitProvider := &mockCommitProvider{}
	mockedDependencyListExporter := &mockDependencyListExporter{}

	mockedProjectProvider.On(
		"GetProject",
		dummyProjectName,
		(*gogitlab.GetProjectOptions)(nil),
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		&gogitlab.Project{
			ID:   dummyProjectID,
			Name: dummyProjectName,
		},
		successGitLabResponse,
		nil,
	)

	mockedBranchProvider.On(
		"GetBranch",
		dummyProjectID,
		dummyBranchName,
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		&gogitlab.Branch{
			Name: dummyBranchName,
			Commit: &gogitlab.Commit{
				ID: dummyCommitSHA,
			},
		},
		successGitLabResponse,
		nil,
	)

	mockedCommitProvider.On(
		"GetCommit",
		dummyProjectID,
		dummyCommitSHA,
		(*gogitlab.GetCommitOptions)(nil),
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		&gogitlab.Commit{
			ID: dummyCommitSHA,
			LastPipeline: &gogitlab.PipelineInfo{
				ID: dummyPipelineID,
			},
		},
		successGitLabResponse,
		nil,
	)

	mockedDependencyListExporter.On(
		"CreateDependencyListExport",
		dummyPipelineID,
		(*gogitlab.CreateDependencyListExportOptions)(nil),
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		expectedCreateDependencyListExport, successGitLabResponse, nil,
	)

	mockedDependencyListExporter.On(
		"GetDependencyListExport",
		dummyExportID,
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		expectedGetDependencyListExport, successGitLabResponse, nil,
	)

	mockedDependencyListExporter.On(
		"DownloadDependencyListExport",
		dummyExportID,
		[]gogitlab.RequestOptionFunc(nil),
	).Return(bytes.NewBuffer(expectedSbomData), successGitLabResponse, nil)

	client := gitlab.NewFetchClient(
		mockedProjectProvider,
		mockedBranchProvider,
		mockedCommitProvider,
		mockedDependencyListExporter,
	)

	glcs.Run("Fetch", func() {
		_, err := client.Fetch(dummyFetchURL, nil)
		glcs.Require().NoError(err, "failed to create dependency list export: %v", err)

		mockedProjectProvider.AssertExpectations(glcs.T())
		mockedBranchProvider.AssertExpectations(glcs.T())
		mockedCommitProvider.AssertExpectations(glcs.T())
		mockedDependencyListExporter.AssertExpectations(glcs.T())
	})
}

func (glcs *gitLabClientSuite) TestClient_Push() {
	dummyHost := "gitlab.dummy"
	dummyProjectID := 1234
	dummyProjectName := "TESTING/TEST"
	dummyPackageName := "SBOM"
	dummyPackageVersion := "1.0.0"

	dummyURL := fmt.Sprintf(
		"https://%s/%s?package_name=%s&package_version=%s",
		dummyHost,
		dummyProjectName,
		dummyPackageName,
		dummyPackageVersion,
	)

	uuidRegex := regexp.MustCompile(`urn:uuid:([\w-]+)`)
	uuidMatch := uuidRegex.FindStringSubmatch(glcs.documents[0].GetMetadata().GetId())
	expectedSbomFileName := uuidMatch[1] + ".json"

	sbomWriter := &gitlab.StringWriter{&strings.Builder{}}
	err := outpututil.WriteStream(glcs.documents[0], "original", glcs.Options, sbomWriter)
	glcs.Require().NoError(err, "failed to serialize SBOM: %v", err)

	expectedSbom := sbomWriter.String()

	mockedProjectProvider := &mockProjectProvider{}
	mockedGenericPackagePublisher := &mockGenericPackagePublisher{}

	mockedProjectProvider.On(
		"GetProject",
		dummyProjectName,
		(*gogitlab.GetProjectOptions)(nil),
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		&gogitlab.Project{
			ID:   dummyProjectID,
			Name: dummyProjectName,
		},
		successGitLabResponse,
		nil,
	)

	mockedGenericPackagePublisher.On(
		"PublishPackageFile",
		dummyProjectID,
		dummyPackageName,
		dummyPackageVersion,
		expectedSbomFileName,
		mock.Anything,
		(*gogitlab.PublishPackageFileOptions)(nil),
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		&gogitlab.GenericPackagesFile{},
		successGitLabResponse,
		nil,
	)

	client := gitlab.NewPushClient(mockedProjectProvider, mockedGenericPackagePublisher)

	glcs.Run("Push", func() {
		err = client.AddFile(
			dummyURL,
			glcs.documents[0].GetMetadata().GetId(),
			&options.PushOptions{
				Options: glcs.Options,
				Format:  "original",
			},
		)
		glcs.Require().NoError(err, "failed to add file: %v", err)
		glcs.Require().Len(client.PushQueue, 1)
		glcs.Require().Equal(expectedSbomFileName, client.PushQueue[0].Name)
		glcs.Require().Equal(expectedSbom, client.PushQueue[0].Contents)

		err = client.Push(dummyURL, nil)
		glcs.Require().NoError(err, "failed to create dependency list export: %v", err)

		mockedProjectProvider.AssertExpectations(glcs.T())
		mockedGenericPackagePublisher.AssertExpectations(glcs.T())
	})
}

func TestGithubClientSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(gitLabClientSuite))
}
