package domain

import "context"

// The actor is the logged in user behind a request. RequireAuth puts it in the request context,
// so code deep in a usecase (e.g. publishing an ActivityLog) knows who acted without every request
// struct carrying the user ID.

type actorContextKey struct{}

// WithActor returns ctx carrying userID as the actor
func WithActor(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, actorContextKey{}, userID)
}

// ActorFrom returns the actor's user ID, or 0 when ctx has none (e.g. a public request)
func ActorFrom(ctx context.Context) int64 {
	id, _ := ctx.Value(actorContextKey{}).(int64)
	return id
}
