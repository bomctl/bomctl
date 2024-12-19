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
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/bomctl/bomctl/internal/pkg/client/gitlab"
)

type (
	gitLabClientSuite struct {
		suite.Suite
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
)

//revive:disable:unchecked-type-assertion

func (mpp *mockProjectProvider) GetProject(
	pid any,
	opt *gogitlab.GetProjectOptions,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.Project, *gogitlab.Response, error) {
	args := mpp.Called(pid, opt, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.Project), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mbp *mockBranchProvider) GetBranch(
	pid any,
	branch string,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.Branch, *gogitlab.Response, error) {
	args := mbp.Called(pid, branch, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.Branch), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mcp *mockCommitProvider) GetCommit(
	pid any,
	sha string,
	opt *gogitlab.GetCommitOptions,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.Commit, *gogitlab.Response, error) {
	args := mcp.Called(pid, sha, opt, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.Commit), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mdle *mockDependencyListExporter) CreateDependencyListExport(
	pipelineID int,
	opt *gogitlab.CreateDependencyListExportOptions,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.DependencyListExport, *gogitlab.Response, error) {
	args := mdle.Called(pipelineID, opt, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.DependencyListExport), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mdle *mockDependencyListExporter) GetDependencyListExport(
	id int,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.DependencyListExport, *gogitlab.Response, error) {
	args := mdle.Called(id, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(*gogitlab.DependencyListExport), args.Get(1).(*gogitlab.Response), args.Error(2)
}

func (mdle *mockDependencyListExporter) DownloadDependencyListExport(
	id int,
	options ...gogitlab.RequestOptionFunc,
) (io.Reader, *gogitlab.Response, error) {
	args := mdle.Called(id, options)

	//nolint:errcheck,wrapcheck
	return args.Get(0).(io.Reader), args.Get(1).(*gogitlab.Response), args.Error(2)
}

//revive:enable:unchecked-type-assertion

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

	client := &gitlab.Client{
		ProjectProvider:        mockedProjectProvider,
		BranchProvider:         mockedBranchProvider,
		CommitProvider:         mockedCommitProvider,
		DependencyListExporter: mockedDependencyListExporter,
		Export:                 nil,
	}

	glcs.Run("Fetch", func() {
		_, err := client.Fetch(dummyFetchURL, nil)
		glcs.Require().NoError(err, "failed to create dependency list export: %v", err)

		mockedProjectProvider.AssertExpectations(glcs.T())
		mockedBranchProvider.AssertExpectations(glcs.T())
		mockedCommitProvider.AssertExpectations(glcs.T())
		mockedDependencyListExporter.AssertExpectations(glcs.T())
	})
}

func TestGithubClientSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(gitLabClientSuite))
}
