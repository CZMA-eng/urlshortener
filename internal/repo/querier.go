// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package repo

import (
	"context"
)

type Querier interface {
	CreateURL(ctx context.Context, arg CreateURLParams) (Url, error)
}

var _ Querier = (*Queries)(nil)
