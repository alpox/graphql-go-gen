package generator

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type Context struct {
	objects  map[string]*graphql.Object
	interfaces map[string]*graphql.Interface
	enums map[string]*graphql.Enum
	unions map[string]*graphql.Union
	scalars map[string]*graphql.Scalar
	inputs map[string]*graphql.InputObject

	processed []int
}

func (g *Context) Object(which string) *graphql.Object {
	return g.objects[which]
}

func (g *Context) Interface(which string) *graphql.Interface {
	return g.interfaces[which]
}

func (g *Context) Enums(which string) *graphql.Enum {
	return g.enums[which]
}

func (g *Context) Union(which string) *graphql.Union {
	return g.unions[which]
}

func (g *Context) Scalar(which string) *graphql.Scalar {
	return g.scalars[which]
}

func (g *Context) InputObject(which string) *graphql.InputObject {
	return g.inputs[which]
}

func (g *Context) GetObject(which string) (graphql.Output, bool) {
	if i, ok := g.scalars[which]; ok {
		return i, true
	}
	if i, ok := g.objects[which]; ok {
		return i, true
	}
	if i, ok := g.interfaces[which]; ok {
		return i, true
	}
	if i, ok := g.unions[which]; ok {
		return i, true
	}
	if i, ok := g.inputs[which]; ok {
		return i, true
	}
	if i, ok := g.enums[which]; ok {
		return i, true
	}
	return nil, false
}

func mapType(ctx *Context, typ ast.Type) (graphql.Output, error) {
	switch typ.(type) {
	case *ast.NonNull:
		nonullType, err := mapType(ctx, typ.(*ast.NonNull).Type)
		if err != nil {
			return nil, err
		}
		return graphql.NewNonNull(nonullType.(graphql.Type)), nil
	case *ast.List:
		listType, err := mapType(ctx, typ.(*ast.List).Type)
		if err != nil {
			return nil, err
		}
		return graphql.NewList(listType.(graphql.Type)), nil
	case *ast.Named:
		switch typ.(*ast.Named).Name.Value {
		case "String":
			return graphql.String, nil
		case "Boolean":
			return graphql.Boolean, nil
		case "Int":
			return graphql.Int, nil
		case "Float":
			return graphql.Float, nil
		case "ID":
			return graphql.ID, nil
		default:
			if ob, ok := ctx.GetObject(typ.(*ast.Named).Name.Value); ok {
				return ob, nil
			}
		}
	}
	return nil, fmt.Errorf("Could not map type %s: Type not found!", typ)
}

func generateFieldArguments(ctx *Context, def *ast.FieldDefinition) (graphql.FieldConfigArgument, error) {
	args := make(graphql.FieldConfigArgument, len(def.Arguments))

	for _, arg := range def.Arguments {
		typ, err := mapType(ctx, arg.Type)
		if err != nil {
			return nil, err
		}

		argConfig := &graphql.ArgumentConfig {
			Type: typ,
		}

		if arg.DefaultValue != nil {
			argConfig.DefaultValue = arg.DefaultValue.GetValue()
		}

		args[arg.Name.Value] = argConfig
	}

	if len(args) > 0 {
		return args, nil
	}
	return nil, nil
}

func generateInputFields(ctx *Context, def *ast.InputObjectDefinition) (graphql.InputObjectConfigFieldMap, error) {
	fields := make(graphql.InputObjectConfigFieldMap, len(def.Fields))
	for _, fieldDef := range def.Fields {
		typ, err := mapType(ctx, fieldDef.Type)
		if err != nil {
			return nil, err
		}

		field := &graphql.InputObjectFieldConfig {
			Type: typ,
		}

		if fieldDef.DefaultValue != nil {
			field.DefaultValue = fieldDef.DefaultValue.GetValue()
		}

		fields[fieldDef.Name.Value] = field
	}

	if len(fields) > 0 {
		return fields, nil
	}
	return nil, nil
}

func generateFields(ctx *Context, def interface{}) (graphql.Fields, error) {
    var fieldDefs []*ast.FieldDefinition

	switch def.(type) {
	case *ast.ObjectDefinition:
		fieldDefs = def.(*ast.ObjectDefinition).Fields
	case *ast.InterfaceDefinition:
		fieldDefs = def.(*ast.InterfaceDefinition).Fields
	default:
		return nil, fmt.Errorf("GenerateFields: Given definition was no Object or Interface definition.")
	}

	fields := make(map[string]*graphql.Field, len(fieldDefs))
	for _, fieldDef := range fieldDefs {
		typ, err := mapType(ctx, fieldDef.Type)
		if err != nil {
			return nil, err
		}

		field :=  &graphql.Field {
			Type: typ,
		}

		args, err := generateFieldArguments(ctx, fieldDef)
		if err != nil {
			return nil, err
		}
		if args != nil {
			field.Args = args
		}

		fields[fieldDef.Name.Value] = field
	}

	if len(fields) > 0 {
		return graphql.Fields(fields), nil
	}
	return nil, nil
}

func generateEnumValues(def *ast.EnumDefinition) graphql.EnumValueConfigMap {
	enumMap := make(graphql.EnumValueConfigMap, len(def.Values))

	for i, valueConfig := range def.Values {
		enumMap[valueConfig.Name.Value] = &graphql.EnumValueConfig {
			Value: i,
		}
	}
	if len(enumMap) > 0 {
		return enumMap
	}
	return nil
}

func generateInterfaces(ctx *Context, obdef *ast.ObjectDefinition) ([]*graphql.Interface, error) {
	ifaces := make([]*graphql.Interface, len(obdef.Interfaces))
	for i, iface := range obdef.Interfaces {
		if lookupIface, ok := ctx.interfaces[iface.Name.Value]; ok {
			ifaces[i] = lookupIface
		} else {
			return nil, fmt.Errorf("An interface with name %s was not declared and can therefore not be " +
			"implemented to object %s\n", iface.Name.Value, obdef.Name.Value)
		}
	}
	if len(ifaces) > 0 {
		return ifaces, nil
	}
	return nil, nil
}

func generateUnionTypes(ctx *Context, def *ast.UnionDefinition) ([]*graphql.Object, error) {
	uTypes := make([]*graphql.Object, len(def.Types))
	for i, utyp := range def.Types {
		if ob, ok := ctx.objects[utyp.Name.Value]; ok {
			uTypes[i] = ob
		} else {
			return nil, fmt.Errorf("An object with name %s was not declared and can therefore not be " +
			"implemented in union %s\n", utyp.Name.Value, def.Name.Value)
		}
	}
	if len(uTypes) > 0 {
		return uTypes, nil
	}
	return nil, nil
}

// Collaborate interfaces on first pass
func walk(context *Context, astDoc *ast.Document) bool {
	var found bool
	for astIndex, def := range astDoc.Definitions {
		var foundInCycle bool
		for _, storedAstIndex := range context.processed  {
			if storedAstIndex == astIndex {
				goto scanNext
			}
		}

		switch def.(type) {
		case *ast.InterfaceDefinition:
			idef := def.(*ast.InterfaceDefinition)

			iConfig := graphql.InterfaceConfig {
				Name: idef.Name.Value,
			}
			fields, err := generateFields(context, idef)
			if err != nil {
				continue // Get in next cycle
			} else if fields != nil {
				iConfig.Fields = fields
			}

			correspondingInterface := graphql.NewInterface(iConfig)
			context.interfaces[idef.Name.Value] = correspondingInterface
			foundInCycle = true
		case *ast.EnumDefinition:
			edef := def.(*ast.EnumDefinition)
			eConfig := graphql.EnumConfig {
				Name: edef.Name.Value,
			}

			values := generateEnumValues(edef)
			if values != nil {
				eConfig.Values = values
			}

			correspondingEnum := graphql.NewEnum(eConfig)
			context.enums[edef.Name.Value] = correspondingEnum
			foundInCycle = true
		case *ast.ScalarDefinition:
			sdef := def.(*ast.ScalarDefinition)
			sConfig := graphql.ScalarConfig {
				Name: sdef.Name.Value,
			}
			correspondingScalar := graphql.NewScalar(sConfig)
			context.scalars[sdef.Name.Value] = correspondingScalar
			foundInCycle = true
		case *ast.UnionDefinition:
			udef := def.(*ast.UnionDefinition)
			uConfig := graphql.UnionConfig {
				Name: udef.Name.Value,
			}

			uTypes, err := generateUnionTypes(context, udef)
			if err != nil {
				continue // Get in next cycle
			}
			if uTypes != nil {
				uConfig.Types = uTypes
			}

			correspondingUnion := graphql.NewUnion(uConfig)
			context.unions[udef.Name.Value] = correspondingUnion
			foundInCycle = true
		case *ast.TypeExtensionDefinition:
			obdef := def.(*ast.TypeExtensionDefinition).Definition
			ob := context.Object(obdef.Name.Value)
			if ob == nil {
				continue // No object with this type. Get in next cycle.
			}
			fields, err := generateFields(context, obdef)
			if err != nil {
				continue
			} else if fields != nil {
				for fieldName, field := range(fields) {
				//	_, ok := ob.Fields()[fieldName]
				//	if ok {
				//		continue // Ignore field since its already implemented
				//	}

					// ** OVERRIDE ** --> Maybe change that behaviour later
					ob.AddFieldConfig(fieldName, field)
				}
			}
			foundInCycle = true
		case *ast.ObjectDefinition:
			obdef := def.(*ast.ObjectDefinition)
			obConfig := graphql.ObjectConfig {
				Name: obdef.Name.Value,
			}

			// Include interfaces
			ifaces, err := generateInterfaces(context, obdef)
			if err != nil {
				continue // Get i next cycle 
			}
			if ifaces != nil {
				obConfig.Interfaces = ifaces
			}

			// Include Fields
			fields, err := generateFields(context, obdef)
			if err != nil {
				continue // Get in next cycle
			} else if fields != nil {
				obConfig.Fields = fields
			}

			correspondingObject := graphql.NewObject(obConfig)
			context.objects[obdef.Name.Value] = correspondingObject
			foundInCycle = true
		case *ast.InputObjectDefinition:
			idef := def.(*ast.InputObjectDefinition)
			iConfig := graphql.InputObjectConfig {
				Name: idef.Name.Value,
			}

			inputFields, err := generateInputFields(context, idef)
			if err != nil {
				continue // Get in next cycle
			} else if inputFields != nil {
				iConfig.Fields = inputFields
			}

			correspondingInput := graphql.NewInputObject(iConfig)
			context.inputs[idef.Name.Value] = correspondingInput
			foundInCycle = true
		}

		if foundInCycle {
			context.processed = append(context.processed, astIndex)
			found = true
		}
scanNext:
	}
	return found
}

func Generate(source string) (*Context, error) {
	astDoc, err := parser.Parse(parser.ParseParams {
		Source: source,
		Options: parser.ParseOptions {
			NoLocation: true,
			NoSource: false,
		},
	})

	if err != nil {
		return nil, err
	}

	context := &Context{}
	context.interfaces = make(map[string]*graphql.Interface)
	context.enums = make(map[string]*graphql.Enum)
	context.scalars = make(map[string]*graphql.Scalar)
	context.inputs = make(map[string]*graphql.InputObject)
	context.unions = make(map[string]*graphql.Union)
	context.objects = make(map[string]*graphql.Object)

	for walk(context, astDoc) { }

	return context, nil
}
