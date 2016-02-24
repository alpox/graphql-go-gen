package generator

import (
	"fmt"
	"testing"
	//"github.com/graphql-go/graphql"
)

func TestSchemaCreation(t *testing.T) {
	gql := `
type Hello {
	f: String
}
type Query {
	a: Hello
}`

	ctx, errp := Generate(gql)
	if errp != nil {
		fmt.Print(errp)
		t.FailNow()
	}
	_, err := CreateSchemaFromContext(ctx)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
}

