package user

import "context"

type userCtx struct{}

func Ctx(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userCtx{}, user)
}

func FromCtx(ctx context.Context) User {
	if value, ok := ctx.Value(userCtx{}).(User); ok {
		return value
	}
	return Anonymous
}
