package generator

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"testing"
	"encoding/json"
)

func TestBasicServer(t *testing.T) {
	schema_string := `
		type Query {
			hello: String	
		}
	`

	schema, serr := Generate(schema_string).Fluent().
		Extend("Query").
			Field("hello").
			Resolve(func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			}).
		Schema()

	if serr != nil {
		fmt.Print(serr)
		t.FailNow()
	}


	// Setup Query
	query := `
	 {
		 hello
	 }
	`

	params := graphql.Params { Schema: schema, RequestString: query }
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		fmt.Print("failed to execute graphql operation, errors: %+v", r.Errors)
		t.FailNow()
	}
	rJSON, _ := json.Marshal(r)
	if string(rJSON) != "{\"data\":{\"hello\":\"world\"}}" {
		fmt.Print("Wrong json delivered by test schema!")
		t.FailNow()
	}
}
