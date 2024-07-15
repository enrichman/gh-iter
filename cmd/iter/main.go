package main

import (
	"context"
	"fmt"
	"time"

	ghiter "github.com/enrichman/gh-iter"
	"github.com/google/go-github/v63/github"
)

func main() {
	client := github.NewClient(nil).WithAuthToken("")

	itIDP := ghiter.NewFromFn(func(ctx context.Context, opts *github.ListCursorOptions) ([]*github.IDPGroup, *github.Response, error) {
		groups, resp, err := client.Teams.ListIDPGroupsInOrganization(ctx, "github", opts)
		return groups.Groups, resp, err
	})

	for group := range itIDP.All() {
		fmt.Printf("Group: %s\n", *group.GroupID)
		time.Sleep(time.Second)
	}

	// itAlerts := ghiter.NewFromFn2(client.Dependabot.ListRepoAlerts, "epinio", "epinio")
	// for alert := range itAlerts.All() {
	// 	fmt.Println(itAlerts.Err())
	// 	fmt.Printf("Alert: %v [after cursor: %v]\n", *alert.Number, itAlerts.Opt.ListOptions.Page)
	// 	time.Sleep(time.Second)
	// }

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 2},
	}
	it := ghiter.NewFromFn1(client.Repositories.ListByOrg, "github").Opts(opt)
	for k := range it.All() {
		fmt.Printf("Repo: %s [current page: %d, next page: %d]\n", *k.Name, opt.Page, it.Raw().NextPage)
		time.Sleep(time.Second)
	}

	// itUsers := ghiter.NewFromFn(client.Users.ListAll).Opts(&github.UserListOptions{Since: 232323, ListOptions: github.ListOptions{PerPage: 2}})
	// for u := range itUsers.All() {
	// 	fmt.Println(itUsers.Err())
	// 	fmt.Println(u)

	// 	fmt.Printf("User: %v [since: %d, next page: %d]\n", *u.ID, itUsers.Opt.Since, itUsers.Raw().NextPage)
	// 	time.Sleep(time.Second)
	// }

	// fmt.Println(itUsers.Err())

	// get all pages of results

}
