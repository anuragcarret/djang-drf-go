package models

import (
	"context"
)

// Lifecycle hook interfaces
type PreSaver interface {
	PreSave(ctx context.Context) error
}

type PostSaver interface {
	PostSave(ctx context.Context, created bool) error
}

type PreDeleter interface {
	PreDelete(ctx context.Context) error
}

type PostDeleter interface {
	PostDelete(ctx context.Context) error
}

type Validator interface {
	Validate() error
}
