package cmdutil

import (
	"context"
	"parm/internal/gh"
)

type ProviderFactory func(ctx context.Context, token string, opts ...gh.Option)

type Factory struct {
	Provider ProviderFactory
}
