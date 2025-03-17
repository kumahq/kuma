package configfetcher

import (
	"context"
	"io"
)

type Handler interface {
	// Path the path to get the config from (Must start with a `/`)
	Path() string
	// OnChange received whenever a new value is received and is different than the last invocation.
	OnChange(ctx context.Context, reader io.Reader) error
	// Shutdown starts and block for shutdown
	Shutdown(ctx context.Context) error
}

type SimpleHandler struct {
	path     string
	onChange func(ctx context.Context, reader io.Reader) error
	shutdown func(ctx context.Context) error
}

func NewSimpleHandler(path string, onChange func(ctx context.Context, reader io.Reader) error, shutdown func(ctx context.Context) error) Handler {
	return &SimpleHandler{
		path:     path,
		onChange: onChange,
		shutdown: shutdown,
	}
}

func (s SimpleHandler) Path() string {
	return s.path
}

func (s SimpleHandler) OnChange(ctx context.Context, reader io.Reader) error {
	return s.onChange(ctx, reader)
}

func (s SimpleHandler) Shutdown(ctx context.Context) error {
	return s.shutdown(ctx)
}
