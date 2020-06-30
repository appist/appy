package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"{{.projectName}}/pkg/graphql/generated"
	"{{.projectName}}/pkg/graphql/model"
	"context"
)

// AddTodo adds a new todo.
func (r *mutationRoot) AddTodo(c context.Context, description, username string) (*model.Todo, error) {
	panic("not implemented")
}

// Mutation returns generated.MutationResolver implementation.
func (r *Root) Mutation() generated.MutationResolver { return &mutationRoot{r} }

type mutationRoot struct{ *Root }
