package profile

import "context"

type Repository interface {
	List(context.Context) ([]Profile, error)
	Get(context.Context, string) (Profile, error)
	Save(context.Context, Profile) error
	Delete(context.Context, string) error
}
