package generator

import (
	"github.com/graphql-go/graphql/language/ast"
)

type mapFunc func(x interface{}) interface{}

func mapSlice(b []ast.Value, fn mapFunc) (newSlice []interface{}) {
	for _, item := range b {
		newSlice = append(newSlice, fn(item))
	}
	return
}
