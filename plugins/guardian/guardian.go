package guardian

import (
	"context"
	"errors"

	"github.com/appist/appy"
)

var (
	// ErrUnsupportedAuthContext indicates the provided auth context is neither appy.Context nor context.Context.
	ErrUnsupportedAuthContext = errors.New("unsupported auth context, should only be appy.Context or context.Context")

	requestCtxKey     = appy.ContextKey("guardian.request")
	currentUserCtxKey = appy.ContextKey("guardian.currentUser")
)

// Init initializes itself to be attached to the appy's application.
func Init(app *appy.App) {
	app.Cmd().AddCommand(newAppyGuardianInstallCommand())
}

// Authenticate checks if the JWT in the session cookie or authorization header is valid and only return 401 HTTP
// status code if the JWT isn't valid or doesn't have match any in the database.
func Authenticate(abort bool) appy.HandlerFunc {
	return func(ctx *appy.Context) {
		newCtx := context.WithValue(ctx.Request.Context(), requestCtxKey, ctx)
		ctx.Request = ctx.Request.WithContext(newCtx)
		ctx.Next()
	}
}

// CurrentUser retrieves the currently logged in user from the context.
func CurrentUser(c interface{}) (interface{}, error) {
	var appyCtx *appy.Context

	switch ctx := c.(type) {
	case *appy.Context:
		appyCtx = ctx
	case context.Context:
		val := ctx.Value(requestCtxKey)
		if val == nil {
			return nil, nil
		}

		appyCtx = val.(*appy.Context)
	default:
		return nil, ErrUnsupportedAuthContext
	}

	currentUser, _ := appyCtx.Get(currentUserCtxKey.String())
	return currentUser, nil
}

// SignIn generates a new session.
func SignIn(ctx *appy.Context) {

}

// SignUp creates a new user record in the database.
func SignUp(ctx *appy.Context) {

}

// SignOut destroy the given session.
func SignOut(ctx *appy.Context) {

}
