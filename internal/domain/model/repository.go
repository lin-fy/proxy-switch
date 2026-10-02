package model

import "context"

type Repository interface {
	List(context.Context) ([]Model, error)
	ListByProvider(context.Context, string) ([]Model, error)
	Get(context.Context, string, string) (Model, error)
	Save(context.Context, Model) error
	Delete(context.Context, string, string) error
}
