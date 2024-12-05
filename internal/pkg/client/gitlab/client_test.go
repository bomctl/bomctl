package gitlab_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	gogitlab "github.com/xanzy/go-gitlab"

	"github.com/bomctl/bomctl/internal/pkg/client/gitlab"
)

type mockClient struct {
	mock.Mock
}

func (mc *mockClient) GetProject(
	pid any,
	opt *gogitlab.GetProjectOptions,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.Project, *gogitlab.Response, error) {
	args := mc.Called(pid, opt, options)

	return args.Get(0).(*gogitlab.Project), args.Get(1).(*gogitlab.Response), args.Error(2) //nolint:errcheck,revive,wrapcheck,lll
}

func (mc *mockClient) GetBranch(
	pid any,
	branch string,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.Branch, *gogitlab.Response, error) {
	args := mc.Called(pid, branch, options)

	return args.Get(0).(*gogitlab.Branch), args.Get(1).(*gogitlab.Response), args.Error(2) //nolint:errcheck,revive,wrapcheck,lll
}

func (mc *mockClient) GetCommit(
	pid any,
	sha string,
	opt *gogitlab.GetCommitOptions,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.Commit, *gogitlab.Response, error) {
	args := mc.Called(pid, sha, opt, options)

	return args.Get(0).(*gogitlab.Commit), args.Get(1).(*gogitlab.Response), args.Error(2) //nolint:errcheck,revive,wrapcheck,lll
}

func (mc *mockClient) CreateDependencyListExport(
	pipelineID int,
	opt *gogitlab.CreateDependencyListExportOptions,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.DependencyListExport, *gogitlab.Response, error) {
	args := mc.Called(pipelineID, opt, options)

	return args.Get(0).(*gogitlab.DependencyListExport), args.Get(1).(*gogitlab.Response), args.Error(2) //nolint:errcheck,revive,wrapcheck,lll
}

func (mc *mockClient) GetDependencyListExport(
	id int,
	options ...gogitlab.RequestOptionFunc,
) (*gogitlab.DependencyListExport, *gogitlab.Response, error) {
	args := mc.Called(id, options)

	return args.Get(0).(*gogitlab.DependencyListExport), args.Get(1).(*gogitlab.Response), args.Error(2) //nolint:errcheck,revive,wrapcheck,lll
}

func (mc *mockClient) DownloadDependencyListExport(
	id int,
	options ...gogitlab.RequestOptionFunc,
) (io.Reader, *gogitlab.Response, error) {
	args := mc.Called(id, options)

	return args.Get(0).(io.Reader), args.Get(1).(*gogitlab.Response), args.Error(2) //nolint:errcheck,revive,wrapcheck
}

var successGitLabResponse = &gogitlab.Response{
	Response: &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.0",
	},
}

func TestFetch(t *testing.T) { //nolint:paralleltest
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

	mockedGoGitLabClient := &mockClient{}

	mockedGoGitLabClient.On(
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

	mockedGoGitLabClient.On(
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

	mockedGoGitLabClient.On(
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

	mockedGoGitLabClient.On(
		"CreateDependencyListExport",
		dummyPipelineID,
		(*gogitlab.CreateDependencyListExportOptions)(nil),
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		expectedCreateDependencyListExport, successGitLabResponse, nil,
	)

	mockedGoGitLabClient.On(
		"GetDependencyListExport",
		dummyExportID,
		[]gogitlab.RequestOptionFunc(nil),
	).Return(
		expectedGetDependencyListExport, successGitLabResponse, nil,
	)

	mockedGoGitLabClient.On(
		"DownloadDependencyListExport",
		dummyExportID,
		[]gogitlab.RequestOptionFunc(nil),
	).Return(bytes.NewBuffer(expectedSbomData), successGitLabResponse, nil)

	_, err := (&gitlab.Client{
		InitFetch: func(c *gitlab.Client) error {
			c.Client = mockedGoGitLabClient
			c.Export = nil

			return nil
		},
	}).Fetch(dummyFetchURL, nil)
	if err != nil {
		t.Errorf("failed to create dependency list export: %v", err)
	}

	mockedGoGitLabClient.AssertExpectations(t)
}
