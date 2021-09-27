package user

import "context"

type userCtx struct{}

func UserCtx(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userCtx{}, &user)
}

func UserFromCtx(ctx context.Context) *User {
	if value, ok := ctx.Value(userCtx{}).(*User); ok {
		return value
	}
	return nil
}
