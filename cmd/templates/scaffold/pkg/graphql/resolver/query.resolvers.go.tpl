package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"{{.projectName}}/pkg/graphql/generated"
	"{{.projectName}}/pkg/graphql/model"
	"context"
)

// Todos returns all the todos.
func (r *queryRoot) Todos(c context.Context) ([]*model.Todo, error) {
	panic("not implemented")
}

// Query returns generated.QueryResolver implementation.
func (r *Root) Query() generated.QueryResolver { return &queryRoot{r} }

type queryRoot struct{ *Root }
