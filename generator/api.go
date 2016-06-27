package generator

import (
	"fmt"
	"reflect"
	"github.com/graphql-go/graphql"
)

type FluentNode struct {
	FluentApi
	ob graphql.Named
	obConfig interface{}
}

type FieldNode struct {
	FluentNode
	fieldConfig *graphql.Field
	pred *FluentNode
}

type FluentApi struct {
	ctx *Context
}

func (f *FluentApi) Context() *Context {
	return f.ctx
}

func (f *FluentApi) Schema() (graphql.Schema, error) {
	return CreateSchemaFromContext(f.ctx)
}

func (f *FluentApi) Extend(which string) *FluentNode {
	ob, ok := f.ctx.GetObject(which)
	if !ok {
		panic(fmt.Sprintf("No object with name %s found!", which))
	}
	config, _ := f.ctx.GetObjectConfig(which)

	return &FluentNode{*f, graphql.GetNamed(ob), config}
}

func (f *FluentNode) Field(fieldName string) *FieldNode {
	fieldType := f.obConfig
	s := reflect.ValueOf(fieldType)
	field := s.FieldByName("Fields").Interface().(graphql.Fields)
	return &FieldNode{*f, field[fieldName], f}
}

func (f *FieldNode) Resolve(fn graphql.FieldResolveFn) *FieldNode {
	f.fieldConfig.Resolve = fn
	return f
}

//func (f *FluentNode) Leave() *FluentApi {
//	return &FluentApi{f.ctx}
//}

//func (f *FieldNode) Leave() *FluentNode {
//	return &FluentNode{f.ctx, f.ob, f.obConfig}
//}
