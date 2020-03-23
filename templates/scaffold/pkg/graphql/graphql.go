package graphql

import (
	"{{.Project.Name}}/pkg/app"
	"{{.Project.Name}}/pkg/graphql/generated"
	"{{.Project.Name}}/pkg/graphql/resolver"

	gqlgen "github.com/99designs/gqlgen/graphql"
)

func init() {
	app.Server.SetupGraphQL(
		"/graphql",
		generated.NewExecutableSchema(newConfig()),
		[]gqlgen.HandlerExtension{},
	)
}

func newConfig() generated.Config {
	return generated.Config{
		Resolvers:  &resolver.Root{},
		Directives: generated.DirectiveRoot{},
	}
}
