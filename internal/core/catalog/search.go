package catalog

import (
	"context"
	"fmt"
	"parm/internal/gh"

	"github.com/google/go-github/v74/github"
)

// TODO: switch to functional/variadic options instead?
type RepoSearchOptions struct {
	Key   *string
	Query *string
}

func SearchRepo(ctx context.Context, provider gh.Provider, opts RepoSearchOptions) (*github.RepositoriesSearchResult, error) {
	search := provider.Search()
	repos := provider.Repos()
	ghOpts := github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
	}
	var query string
	if opts.Key != nil {
		// TODO: change this query logic to something better
		query = fmt.Sprintf("q=%s", *opts.Key)
	} else if opts.Query != nil {
		query = *opts.Query
	} else {
		// both null, return err
		return nil, fmt.Errorf("error: query cannot be nil")
	}
	res, _, err := search.Repositories(ctx, query, &ghOpts)
	if err != nil {
		return nil, fmt.Errorf("error: could not search repositories:\n%q", err)
	}

	// Filter for only repos that have releases
	var filteredRepos []*github.Repository
	// Limit to top 10 results to avoid excessive API calls
	maxToProcess := len(res.Repositories)
	if maxToProcess > 10 {
		maxToProcess = 10
	}

	for i := 0; i < maxToProcess; i++ {
		repo := res.Repositories[i]
		owner := repo.GetOwner().GetLogin()
		name := repo.GetName()

		// Check for at least one release
		rel, _, err := repos.GetLatestRelease(ctx, owner, name)
		if err == nil && rel != nil {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	res.Repositories = filteredRepos
	res.Total = github.Ptr(len(filteredRepos))

	return res, nil
}
