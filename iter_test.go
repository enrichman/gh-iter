package ghiter

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v64/github"
)

func Test_updateOptions(t *testing.T) {
	tt := []struct {
		name         string
		opts         any
		queryParams  map[string]string
		expectedOpts any
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
	}

	for _, tc := range tt {
		updateOptions(tc.opts, tc.queryParams)
		if !reflect.DeepEqual(tc.expectedOpts, tc.opts) {
			t.Fatalf("structs are not equal:\nexpected:\t%+v\ngot:\t\t%+v\n", tc.expectedOpts, tc.opts)
		}
	}
}
