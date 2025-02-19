package ghiter

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
)

func Test_updateOptions(t *testing.T) {
	tt := []struct {
		name         string
		opts         any
		queryParams  map[string]string
		expectedOpts any
		expectedErr  bool
	}{
		{
			name: "Simple Opts with ListOptions",
			opts: &github.RepositoryListByOrgOptions{},
			queryParams: map[string]string{
				"page": "1123323",
			},
			expectedOpts: &github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{
					Page: 1123323,
				},
			},
		},
		{
			name: "Simple Opts with ListOptions with multiple params",
			opts: &github.RepositoryListByOrgOptions{},
			queryParams: map[string]string{
				"page":      "1123323",
				"direction": "desc",
			},
			expectedOpts: &github.RepositoryListByOrgOptions{
				Direction: "desc",
				ListOptions: github.ListOptions{
					Page: 1123323,
				},
			},
		},
		{
			name: "Opts with Since",
			opts: &github.RepositoryListAllOptions{},
			queryParams: map[string]string{
				"page":      "1123323",
				"direction": "desc",
				"since":     "111",
			},
			expectedOpts: &github.RepositoryListAllOptions{
				Since: 111,
			},
		},
		{
			name: "Opts with pointers and multiple ListOptions",
			opts: &github.ListAlertsOptions{},
			queryParams: map[string]string{
				"page":      "1123323",
				"direction": "desc",
				"since":     "111",
				"sort":      "",
			},
			expectedOpts: &github.ListAlertsOptions{
				Direction: func() *string {
					direction := "desc"
					return &direction
				}(),
				Sort: func() *string {
					var sort string
					return &sort
				}(),
				ListOptions: github.ListOptions{
					Page: 1123323,
				},
				ListCursorOptions: github.ListCursorOptions{
					Page: "1123323",
				},
			},
		},
		{
			name: "date RFC parse",
			opts: &github.IssueListCommentsOptions{},
			queryParams: map[string]string{
				"since": "1989-10-02T00:00:00Z",
			},
			expectedOpts: &github.IssueListCommentsOptions{
				Since: func() *time.Time {
					since := time.Date(1989, time.October, 2, 0, 0, 0, 0, time.UTC)
					return &since
				}(),
			},
		},
		{
			name: "date DateTime parse",
			opts: &github.IssueListCommentsOptions{},
			queryParams: map[string]string{
				"since": "1989-10-02",
			},
			expectedOpts: &github.IssueListCommentsOptions{
				Since: func() *time.Time {
					since := time.Date(1989, time.October, 2, 0, 0, 0, 0, time.UTC)
					return &since
				}(),
			},
		},
		{
			name: "wrong Date",
			opts: &github.IssueListCommentsOptions{},
			queryParams: map[string]string{
				"since": "1923230Z",
			},
			expectedErr: true,
		},
		{
			name: "Opts with pointers and multiple ListOptions",
			opts: &github.CommitsListOptions{},
			queryParams: map[string]string{
				"since": "1989-10-02T00:00:00Z",
			},
			expectedOpts: &github.CommitsListOptions{
				Since: time.Date(1989, time.October, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Opts with bool",
			opts: &github.ListWorkflowRunsOptions{},
			queryParams: map[string]string{
				"exclude_pull_requests": "true",
			},
			expectedOpts: &github.ListWorkflowRunsOptions{
				ExcludePullRequests: true,
			},
		},
		{
			name: "Opts with bool pointers",
			opts: &github.WorkflowRunAttemptOptions{},
			queryParams: map[string]string{
				"exclude_pull_requests": "true",
			},
			expectedOpts: &github.WorkflowRunAttemptOptions{
				ExcludePullRequests: func() *bool {
					exclude := true
					return &exclude
				}(),
			},
		},
		{
			name: "Opts with int pointers",
			opts: &github.ListSCIMProvisionedIdentitiesOptions{},
			queryParams: map[string]string{
				"count": "12",
			},
			expectedOpts: &github.ListSCIMProvisionedIdentitiesOptions{
				Count: func() *int {
					count := 12
					return &count
				}(),
			},
		},
		{
			name: "Opts with int64 pointers",
			opts: &github.ListCheckRunsOptions{},
			queryParams: map[string]string{
				"app_id": "12",
			},
			expectedOpts: &github.ListCheckRunsOptions{
				AppID: func() *int64 {
					count := int64(12)
					return &count
				}(),
			},
		},
	}

	for _, tc := range tt {
		err := updateOptions(tc.opts, tc.queryParams)

		if tc.expectedErr {
			if err == nil {
				t.Fatal("missing expected err\n\n")
			}
			continue
		}

		if !reflect.DeepEqual(tc.expectedOpts, tc.opts) {
			t.Fatalf("structs are not equal:\nexpected:\t%+v\ngot:\t\t%+v\n", tc.expectedOpts, tc.opts)
		}
	}
}
