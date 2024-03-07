package db

import "context"

type ISeeder interface {
	GetName() string
	Handle(ctx context.Context) error
}
