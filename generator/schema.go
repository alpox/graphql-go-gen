package generator

import (
	"errors"
	"github.com/graphql-go/graphql"
)

func CreateSchemaFromContext(ctx *Context) (graphql.Schema, error) {
	if query, ok := ctx.objects["Query"]; ok {
		return graphql.NewSchema(graphql.SchemaConfig {
			Query: query,
		})
	} else {
		return graphql.Schema{}, errors.New("Your context does not define a Query root type!")
	}
}
