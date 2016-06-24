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

	ctx, err := Generate(schema_string)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}

	ctx.Extend("Query", UpdateObjectFn(func(config graphql.ObjectConfig) graphql.ObjectConfig {
		fields := config.Fields.(graphql.Fields)
		fields["hello"].Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			return "world", nil
		}
		return config
	}))

	schema, serr := CreateSchemaFromContext(ctx)
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
	fmt.Printf("%s \n", rJSON)
}
