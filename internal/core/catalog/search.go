package catalog

import (
	"context"
	"fmt"

	"github.com/google/go-github/v74/github"
)

// TODO: switch to functional/variadic options instead?
type RepoSearchOptions struct {
	Key   *string
	Query *string
}

func SearchRepo(ctx context.Context, search *github.SearchService, opts RepoSearchOptions) (*github.RepositoriesSearchResult, error) {
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
		return nil, fmt.Errorf("query cannot be nil")
	}
	res, _, err := search.Repositories(ctx, query, &ghOpts)
	if err != nil {
		return nil, fmt.Errorf("could not search repositories:\n%q", err)
	}

	// TODO: filter for only repos that have releases
	// var repos [][2]string
	//
	// for _, repo := range res.Repositories {
	// }

	return res, nil
}
