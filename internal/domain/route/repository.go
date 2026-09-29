package route

import "context"

type Repository interface {
	List(context.Context) ([]Route, error)
	Get(context.Context, string) (Route, error)
	Save(context.Context, Route) error
	Delete(context.Context, string) error
}
