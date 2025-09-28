package context

import (
	"context"

	"github.com/Rahul4469/lenslocked/models"
)

type key string

const (
	userKey key = "user"
)

// Takes built-in context type and User type,
// binds the User objext/data to ctx and return the data
// as ctx of context type to use later
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// Takes the ctx of context type with attached user data
// and returns *models.User object
func User(ctx context.Context) *models.User {
	val := ctx.Value(userKey)
	user, ok := val.(*models.User)
	if !ok {
		return nil
	}
	return user
}
