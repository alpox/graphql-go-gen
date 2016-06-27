package generator

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"strconv"
	"log"
)

type UpdateObjectFn      func(graphql.ObjectConfig) graphql.ObjectConfig
type UpdateInterfaceFn	 func(graphql.InterfaceConfig) graphql.InterfaceConfig
type UpdateEnumFn		 func(graphql.EnumConfig) graphql.EnumConfig
type UpdateUnionFn		 func(graphql.UnionConfig) graphql.UnionConfig
type UpdateScalarFn		 func(graphql.ScalarConfig) graphql.ScalarConfig
type UpdateInputObjectFn func(graphql.InputObjectConfig) graphql.InputObjectConfig

type Context struct {
	objects    map[string]*graphql.Object
	interfaces map[string]*graphql.Interface
	enums      map[string]*graphql.Enum
	unions     map[string]*graphql.Union
	scalars    map[string]*graphql.Scalar
	inputs     map[string]*graphql.InputObject

	objectConfigs    map[string]graphql.ObjectConfig
	interfaceConfigs map[string]graphql.InterfaceConfig
	enumConfigs      map[string]graphql.EnumConfig
	unionConfigs     map[string]graphql.UnionConfig
	scalarConfigs    map[string]graphql.ScalarConfig
	inputConfigs     map[string]graphql.InputObjectConfig

	processed []int
}

func (g *Context) Fluent() *FluentApi {
	return &FluentApi{g}
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

func (g *Context) GetObjectConfig(which string) (interface{}, bool) {
	if i, ok := g.scalarConfigs[which]; ok {
		return i, true
	}
	if i, ok := g.objectConfigs[which]; ok {
		return i, true
	}
	if i, ok := g.interfaceConfigs[which]; ok {
		return i, true
	}
	if i, ok := g.unionConfigs[which]; ok {
		return i, true
	}
	if i, ok := g.inputConfigs[which]; ok {
		return i, true
	}
	if i, ok := g.enumConfigs[which]; ok {
		return i, true
	}
	return nil, false
}

func (g *Context) UpdateObject(which string, config interface{}) error {
	config, ok := g.GetObjectConfig(which)
	if !ok {
		return fmt.Errorf("Could not find Object with name %s.\n", which)
	}
	if _, ok := g.scalars[which]; ok {
		g.scalars[which] = graphql.NewScalar(config.(graphql.ScalarConfig))
	}
	if _, ok := g.objects[which]; ok {
		g.objects[which] = graphql.NewObject(config.(graphql.ObjectConfig))
	}
	if _, ok := g.interfaces[which]; ok {
		g.interfaces[which] = graphql.NewInterface(config.(graphql.InterfaceConfig))
	}
	if _, ok := g.unions[which]; ok {
		g.unions[which] = graphql.NewUnion(config.(graphql.UnionConfig))
	}
	if _, ok := g.inputs[which]; ok {
		g.inputs[which] = graphql.NewInputObject(config.(graphql.InputObjectConfig))
	}
	if _, ok := g.enums[which]; ok {
		g.enums[which] = graphql.NewEnum(config.(graphql.EnumConfig))
	}
	return nil
}

func (g *Context) Extend(which string, updateFn interface{}) *Context {
	objectConfig, ok := g.GetObjectConfig(which)
	if !ok {
		panic("No object with name " + which + " found.\n")
	}

	switch objectConfig.(type) {
	case graphql.ObjectConfig:
		fn, ok := updateFn.(UpdateObjectFn)
		if !ok {
			panic("Given updatefunction (updateFn) has to be of type UpdateObjectFn!")
		}

		obConfig, ok := objectConfig.(graphql.ObjectConfig)
		if !ok {
			panic("Object found is not of type graphql.Object!\n")
		}

		fn(obConfig)
		break
	case graphql.InterfaceConfig:
		fn, ok := updateFn.(UpdateInterfaceFn)
		if !ok {
			panic("Given updatefunction (updateFn) has to be of type UpdateObjectFn!")
		}

		obConfig, ok := objectConfig.(graphql.InterfaceConfig)
		if !ok {
			panic("Object found is not of type graphql.Object!\n")
		}

		fn(obConfig)
		break
	case graphql.UnionConfig:
		fn, ok := updateFn.(UpdateUnionFn)
		if !ok {
			panic("Given updatefunction (updateFn) has to be of type UpdateUnionFn!")
		}

		obConfig, ok := objectConfig.(graphql.UnionConfig)
		if !ok {
			panic("Object found is not of type graphql.Union!")
		}

		fn(obConfig)
		break
	case graphql.ScalarConfig:
		fn, ok := updateFn.(UpdateScalarFn)
		if !ok {
			panic("Given updatefunction (updateFn) has to be of type UpdateScalarFn!")
		}

		obConfig, ok := objectConfig.(graphql.ScalarConfig)
		if !ok {
			panic("Object found is not of type graphql.Scalar!")
		}

		fn(obConfig)
		break
	case graphql.EnumConfig:
		fn, ok := updateFn.(UpdateEnumFn)
		if !ok {
			panic("Given updatefunction (updateFn) has to be of type UpdateEnumFn!")
		}

		obConfig, ok := objectConfig.(graphql.EnumConfig)
		if !ok {
			panic("Object found is not of type graphql.Enum!")
		}

		fn(obConfig)
		break
	case graphql.InputObjectConfig:
		fn, ok := updateFn.(UpdateInputObjectFn)
		if !ok {
			panic("Given updatefunction (updateFn) has to be of type UpdateInputObjectFn!")
		}

		obConfig, ok := objectConfig.(graphql.InputObjectConfig)
		if !ok {
			panic("Object found is not of type graphql.InputObject!")
		}

		fn(obConfig)
	}
	return g
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

func createValues(valType interface{}) interface{} {
	value := valType.(ast.Value)
	return value.GetValue()
}

func mapValue(ctx *Context, value interface{}) (interface{}, error) {
	switch value.(type) {
	case *ast.ListValue:
		listValue := value.(*ast.ListValue)
		return mapSlice(listValue.Values, createValues), nil
	case *ast.FloatValue:
		val := value.(*ast.FloatValue)
		return strconv.ParseFloat(val.Value, 64)
	case *ast.IntValue:
		val := value.(*ast.IntValue)
		return strconv.Atoi(val.Value)
	case *ast.StringValue:
		val := value.(*ast.StringValue)
		return val.Value, nil
	case *ast.BooleanValue:
		val := value.(*ast.BooleanValue)
		return val.Value, nil
	}
	return nil, nil
}

func generateFieldArguments(ctx *Context, def *ast.FieldDefinition) (graphql.FieldConfigArgument, error) {
	args := make(graphql.FieldConfigArgument, len(def.Arguments))

	for _, arg := range def.Arguments {
		typ, err := mapType(ctx, arg.Type)
		if err != nil {
			return nil, err
		}

		argConfig := &graphql.ArgumentConfig{
			Type: typ,
		}

		if arg.DefaultValue != nil {
			defaultValue, mapErr := mapValue(ctx, arg.DefaultValue)
			if err != nil {
				return nil, mapErr
			}
			argConfig.DefaultValue = defaultValue
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

		field := &graphql.InputObjectFieldConfig{
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

		field := &graphql.Field{
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
		enumMap[valueConfig.Name.Value] = &graphql.EnumValueConfig{
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
			return nil, fmt.Errorf("An interface with name %s was not declared and can therefore not be "+
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
			return nil, fmt.Errorf("An object with name %s was not declared and can therefore not be "+
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
		for _, storedAstIndex := range context.processed {
			if storedAstIndex == astIndex {
				goto scanNext
			}
		}

		switch def.(type) {
		case *ast.InterfaceDefinition:
			idef := def.(*ast.InterfaceDefinition)

			iConfig := graphql.InterfaceConfig{
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
			context.interfaceConfigs[idef.Name.Value] = iConfig
			foundInCycle = true
		case *ast.EnumDefinition:
			edef := def.(*ast.EnumDefinition)
			eConfig := graphql.EnumConfig{
				Name: edef.Name.Value,
			}

			values := generateEnumValues(edef)
			if values != nil {
				eConfig.Values = values
			}

			correspondingEnum := graphql.NewEnum(eConfig)
			context.enums[edef.Name.Value] = correspondingEnum
			context.enumConfigs[edef.Name.Value] = eConfig
			foundInCycle = true
		case *ast.ScalarDefinition:
			sdef := def.(*ast.ScalarDefinition)
			sConfig := graphql.ScalarConfig{
				Name: sdef.Name.Value,
			}
			correspondingScalar := graphql.NewScalar(sConfig)
			context.scalars[sdef.Name.Value] = correspondingScalar
			context.scalarConfigs[sdef.Name.Value] = sConfig
			foundInCycle = true
		case *ast.UnionDefinition:
			udef := def.(*ast.UnionDefinition)
			uConfig := graphql.UnionConfig{
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
			context.unionConfigs[udef.Name.Value] = uConfig
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
				for fieldName, field := range fields {
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
			obConfig := graphql.ObjectConfig{
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
			context.objectConfigs[obdef.Name.Value] = obConfig
			foundInCycle = true
		case *ast.InputObjectDefinition:
			idef := def.(*ast.InputObjectDefinition)
			iConfig := graphql.InputObjectConfig{
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
			context.inputConfigs[idef.Name.Value] = iConfig
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

func Generate(source string) *Context {
	astDoc, err := parser.Parse(parser.ParseParams{
		Source: source,
		Options: parser.ParseOptions{
			NoLocation: true,
			NoSource:   false,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	context := &Context{}
	context.interfaces = make(map[string]*graphql.Interface)
	context.enums = make(map[string]*graphql.Enum)
	context.scalars = make(map[string]*graphql.Scalar)
	context.inputs = make(map[string]*graphql.InputObject)
	context.unions = make(map[string]*graphql.Union)
	context.objects = make(map[string]*graphql.Object)

	context.interfaceConfigs = make(map[string]graphql.InterfaceConfig)
	context.enumConfigs = make(map[string]graphql.EnumConfig)
	context.scalarConfigs = make(map[string]graphql.ScalarConfig)
	context.inputConfigs = make(map[string]graphql.InputObjectConfig)
	context.unionConfigs = make(map[string]graphql.UnionConfig)
	context.objectConfigs = make(map[string]graphql.ObjectConfig)

	for walk(context, astDoc) {
	}

	return context
}
