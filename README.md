# gh-iter

> **Note:** this package leverages the new [`iter`](https://pkg.go.dev/iter) package, and it needs Go [`1.23`](https://go.dev/dl/#go1.23).

The `gh-iter` package provides an iterator that can be used with the [`google/go-github`](https://github.com/google/go-github) client.  
It supports automatic pagination with generic types.

## Quickstart

```go
package main

import (
	"fmt"

	ghiter "github.com/enrichman/gh-iter"
	"github.com/google/go-github/v68/github"
)

func main() {
	// init your Github client
	client := github.NewClient(nil)

	// create an iterator, and start looping! 🎉
	users := ghiter.NewFromFn(client.Users.ListAll)
	for u := range users.All() {
		fmt.Println(*u.Login)
	}

	// check if the loop stopped because of an error
	if err := users.Err(); err != nil {
		// something happened :(
		panic(err)
	}
}
```

## Usage

Depending of the API you need to use you can create an iterator from one of the three provided constructor:

### No args

```go
ghiter.NewFromFn(client.Users.ListAll)
```

### One string arg

```go
ghiter.NewFromFn1(client.Repositories.ListByUser, "enrichman")
```

### Two string args

```go
ghiter.NewFromFn2(client.Issues.ListByRepo, "enrichman", "gh-iter")
```

Then you can simply loop through the objects with the `All()` method.


### Customize options

You can tweak the iteration providing your own options. They will be updated during the loop.

For example if you want to request only 5 repositories per request:

```go
ghiter.NewFromFn1(client.Repositories.ListByUser, "enrichman").
	Opts(&github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: 5},
	})
```

### Context

If you don't provide a context with the `Ctx()` func an empty `context.Background` will be used. You can use a custom context to have a more granular control, for example if you want to close the iteration from a timeout, or with a manual cancellation.

You can check if the int


```go
ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

repos := ghiter.NewFromFn1(client.Repositories.ListByUser, "enrichman").Ctx(ctx)
for repo := range repos.All() {
    if *repo.Name == "myrepo" {
        fmt.Println(*repo.Name)
	    cancel()
    }
}
```

## Advanced usage

Some APIs do not match the "standard" string arguments, or the returned type is not an array. In these cases you can still use this package, but you will need to provide a "custom func" to the `ghiter.NewFromFn` constructor.

For example the [`client.Teams.ListTeamReposByID`](https://pkg.go.dev/github.com/google/go-github/v68/github#TeamsService.ListTeamReposByID) needs the `orgID, teamID int64` arguments:

```go
repos := ghiter.NewFromFn(func(ctx context.Context, opts *github.ListOptions) ([]*github.Repository, *github.Response, error) {
	return client.Teams.ListTeamReposByID(ctx, 123, 456, opts)
})
```

In case the returned object is not an array you will have to "unwrap" it.  
For example the [`client.Teams.ListIDPGroupsInOrganization`](https://pkg.go.dev/github.com/google/go-github/v68/github#TeamsService.ListIDPGroupsInOrganization) returns a [IDPGroupList](https://pkg.go.dev/github.com/google/go-github/v68/github#IDPGroupList), and not a slice. 

```go
idpGroups := ghiter.NewFromFn(func(ctx context.Context, opts *github.ListCursorOptions) ([]*github.IDPGroup, *github.Response, error) {
	groups, resp, err := client.Teams.ListIDPGroupsInOrganization(ctx, "myorg", opts)
    // remember to check for nil!
	if groups != nil {
		return groups.Groups, resp, err
	}
	return nil, resp, err
})
```


# Feedback

If you like the project please star it on Github 🌟, and feel free to drop me a note, or [open an issue](https://github.com/enrichman/gh-iter/issues/new)!

[Twitter](https://twitter.com/enrichmann)

# License

[MIT](LICENSE)
