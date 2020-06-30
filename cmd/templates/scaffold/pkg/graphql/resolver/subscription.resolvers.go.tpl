package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"{{.projectName}}/pkg/graphql/generated"
	"{{.projectName}}/pkg/graphql/model"
	"context"
)

// TodoAdded broadcast the message when a new todo is added.
func (r *subscriptionRoot) TodoAdded(c context.Context, channel string) (<-chan *model.Todo, error) {
	panic("not implemented")
}

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Root) Subscription() generated.SubscriptionResolver { return &subscriptionRoot{r} }

type subscriptionRoot struct{ *Root }
