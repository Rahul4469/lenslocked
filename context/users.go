package context

import (
	"context"

	"github.com/Rahul4469/lenslocked/models"
)

type key string

const (
	userKey key = "user"
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return nil
}
