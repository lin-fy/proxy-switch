package provider

import "context"

type Repository interface {
	List(context.Context) ([]Provider, error)
	Get(context.Context, string) (Provider, error)
	Save(context.Context, Provider) error
	Delete(context.Context, string) error
}
