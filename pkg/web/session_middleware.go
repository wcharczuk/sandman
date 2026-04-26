package web

// SessionAware is an action that injects the session into the context.
func SessionAware(action Action) Action {
	return func(ctx Context) Result {
		session, err := ctx.App().VerifyOrExtendSession(ctx)
		if err != nil && !IsErrSessionInvalid(err) {
			return AcceptedProvider(ctx).InternalError(err)
		}
		return action(ctx.WithSession(session))
	}
}

// SessionAwareStable is an action that injects the session into the context, but does
// not extend it if there is a session lifetime handler on the auth manager.
//
// It is typically used for logout endpoints, where you don't want a session
// extension `Set-Cookie` response header to compete with the expiry header.
func SessionAwareStable(action Action) Action {
	return func(ctx Context) Result {
		_, session, err := ctx.App().VerifySession(ctx)
		if err != nil && !IsErrSessionInvalid(err) {
			return AcceptedProvider(ctx).InternalError(err)
		}
		return action(ctx.WithSession(session))
	}
}

// SessionRequired is an action that requires a session to be present
// or identified in some form on the request.
func SessionRequired(action Action) Action {
	return func(ctx Context) Result {
		session, err := ctx.App().VerifyOrExtendSession(ctx)
		if err != nil && !IsErrSessionInvalid(err) {
			return AcceptedProvider(ctx).InternalError(err)
		}
		if session == nil {
			return ctx.App().LoginRedirect(ctx)
		}
		return action(ctx.WithSession(session))
	}
}
