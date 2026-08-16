package middleware

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type authStateKey struct{}

type AuthState int

const (
	AuthOK AuthState = iota
)

func HasAuthState(want AuthState) th.Predicate {
	return func(ctx context.Context, update telego.Update) bool {
		got, ok := ctx.Value(authStateKey{}).(AuthState)

		return ok && got == want
	}
}
