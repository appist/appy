package resolver

import (
	"{{.projectName}}/pkg/graphql/generated"
	"{{.projectName}}/pkg/graphql/model"
	"context"
)

// Root is the entry for GraphQL resolving.
type Root struct{}

// Mutation defines the GraphQL mutations in `pkg/graphql/resolver/mutation`.
func (r *Root) Mutation() generated.MutationResolver {
	return &Mutation{r}
}

// Query defines the GraphQL queries in `pkg/graphql/resolver/query`.
func (r *Root) Query() generated.QueryResolver {
	return &Query{r}
}

// Subscription defines the GraphQL queries in `pkg/graphql/resolver/subscription`.
func (r *Root) Subscription() generated.SubscriptionResolver {
	return &Subscription{r}
}

// Mutation is the entry for all GraphQL mutations.
type Mutation struct{ *Root }

// AddTodo adds a new todo.
func (r *Mutation) AddTodo(c context.Context, description, username string) (*model.Todo, error) {
	panic("not implemented")
}

// Query is the entry for all GraphQL queries.
type Query struct{ *Root }

// Todos returns all the todos.
func (r *Query) Todos(c context.Context) ([]*model.Todo, error) {
	panic("not implemented")
}

// Subscription is the entry for all GraphQL subscriptions.
type Subscription struct{ *Root }

// TodoAdded broadcast the message when a new todo is added.
func (r *Subscription) TodoAdded(c context.Context, channel string) (<-chan *model.Todo, error) {
	panic("not implemented")
}
