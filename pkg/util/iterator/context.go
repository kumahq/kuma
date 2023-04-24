package iterator

import (
	"context"
	"github.com/pkg/errors"
)

var CustomIterator = func(parent ...context.Context) ([]context.Context, error) {
	switch len(parent) {
	case 1:
		return []context.Context{parent[0]}, nil
	case 0:
		return []context.Context{context.Background()}, nil
	default:
		return nil, errors.New("there can only be one parent")
	}
}
