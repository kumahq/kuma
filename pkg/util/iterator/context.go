package iterator

import "context"

var CustomIterator = func() ([]context.Context, error) {
	return []context.Context{context.Background()}, nil
}
