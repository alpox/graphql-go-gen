package generator

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/graphql-go/graphql"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func matchStrings(exp, got string) bool {
	trexp := strings.TrimSpace(exp)
	trgot := strings.TrimSpace(got)
	reg := regexp.MustCompile(`0x[0-9a-z]+`)
	trexp = reg.ReplaceAllString(trexp, "")
	trgot = reg.ReplaceAllString(trgot, "")
	return trexp == trgot
}

func nextLineMatch(exp, got []string, index1, index2 *int) bool {
	for i := *index1 + 1; i < len(exp); i++ {
		for p := *index2 + 1; p < len(got); p++ {
			if matchStrings(exp[i], got[p]) {
				*index1 = i
				*index2 = p
				return true
			}
		}
	}
	*index1, *index2 = -1, -1
	return false
}

func compareLines(str1, str2 string) {
	l1 := strings.Split(str1, "\n")
	l2 := strings.Split(str2, "\n")

	index1 := -1
	index2 := -1
	lastExp := -1
	lastGot := -1
	for nextLineMatch(l1, l2, &index1, &index2) {
		if index1-lastExp > 1 || index2-lastGot > 1 {
			// Unmatching lines in between!
			unmatched := l1[lastExp+1 : index1]
			unmatchedGot := l2[lastGot+1 : index2]
			fmt.Println("-----------------------------------------------")
			fmt.Println("Mismatch. Expected:")
			fmt.Println(strings.Join(unmatched, "\n"))
			fmt.Println("Got:")
			fmt.Println(strings.Join(unmatchedGot, "\n"))
			fmt.Println("-----------------------------------------------")
		}
		lastExp = index1
		lastGot = index2
	}

	if len(l1)-1 > lastExp {
		unmatched := l1[lastExp+1 : len(l1)-1]
		unmatchedGot := l2[lastGot+1 : len(l2)-1]
		fmt.Println("Mismatch. Expected:\n")
		fmt.Println(strings.Join(unmatched, "\n"))
		fmt.Println("\nGot:\n")
		fmt.Println(strings.Join(unmatchedGot, "\n"))
	}
}

func printFail(var1, var2 interface{}, t *testing.T) {
	t.Fail()
	spew.Config.DisableMethods = true
	fmt.Println("Unexpected schema found.\n")
	compareLines(spew.Sdump(var1), spew.Sdump(var2))
	//compareVars("", var1, var2)
}

func TestSimpleType(t *testing.T) {
	gql := `
type Oncle {
	pipe: ID
	five(argument: [String] = ["String", "String"] ): String
}
	`

	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Oncle",
		Fields: graphql.Fields{
			"pipe": &graphql.Field{
				Type: graphql.ID,
			},
			"five": &graphql.Field{
				Type: graphql.String,
				Args: graphql.FieldConfigArgument{
					"argument": &graphql.ArgumentConfig{
						Type:         graphql.NewList(graphql.String),
						DefaultValue: []interface{}{"String", "String"},
					},
				},
			},
		},
	})

	ctx, _ := Generate(gql)
	oncle := ctx.Object("Oncle")
	if !reflect.DeepEqual(oncle, expected) {
		printFail(expected, oncle, t)
	}
}

func TestReferenceTypes(t *testing.T) {
	gql := `
type Oncle {
	pipe: p 
}
type p {
	feld: e
}
interface e {}
	`

	e := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "e",
	})
	p := graphql.NewObject(graphql.ObjectConfig{
		Name: "p",
		Fields: graphql.Fields{
			"feld": &graphql.Field{
				Type: e,
			},
		},
	})
	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Oncle",
		Fields: graphql.Fields{
			"pipe": &graphql.Field{
				Type: p,
			},
		},
	})

	ctx, _ := Generate(gql)
	oncle := ctx.Object("Oncle")
	if !reflect.DeepEqual(oncle, expected) {
		printFail(expected, oncle, t)
	}
}

func TestRequiredType(t *testing.T) {
	gql := `
type Oncle {
	pipe: String!
}
	`

	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Oncle",
		Fields: graphql.Fields{
			"pipe": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	ctx, _ := Generate(gql)
	oncle := ctx.Object("Oncle")
	if !reflect.DeepEqual(oncle, expected) {
		printFail(expected, oncle, t)
	}
}

func TestSimpleInterface(t *testing.T) {
	gql := `
interface World {}
`
	expected := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "World",
	})

	ctx, _ := Generate(gql)
	world := ctx.Interface("World")
	if !reflect.DeepEqual(world, expected) {
		printFail(expected, world, t)
	}
}

func TestSimpleTypeImplementsInterface(t *testing.T) {
	gql := `
interface World {}
type Oncle implements World {}
`
	world := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "World",
	})
	expected := graphql.NewObject(graphql.ObjectConfig{
		Name:       "Oncle",
		Interfaces: []*graphql.Interface{world},
	})

	ctx, _ := Generate(gql)
	oncle := ctx.Object("Oncle")
	if !reflect.DeepEqual(oncle, expected) {
		printFail(expected, oncle, t)
	}
}

func TestSimpleTypeImplementsMultipleInterfaces(t *testing.T) {
	gql := `
interface World {}
interface Balloon {}
type Oncle implements World, Balloon {}
`
	world := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "World",
	})
	balloon := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "Balloon",
	})
	expected := graphql.NewObject(graphql.ObjectConfig{
		Name:       "Oncle",
		Interfaces: []*graphql.Interface{world, balloon},
	})

	ctx, _ := Generate(gql)
	oncle := ctx.Object("Oncle")
	if !reflect.DeepEqual(oncle, expected) {
		printFail(expected, oncle, t)
	}
}

func TestSingleValueEnum(t *testing.T) {
	gql := `enum Hello { WORLD }`

	expected := graphql.NewEnum(graphql.EnumConfig{
		Name: "Hello",
		Values: graphql.EnumValueConfigMap{
			"WORLD": &graphql.EnumValueConfig{
				Value: 0,
			},
		},
	})

	ctx, _ := Generate(gql)
	enum := ctx.Enums("Hello")
	if !reflect.DeepEqual(enum, expected) {
		printFail(expected, enum, t)
	}
}

func TestMultiValueEnum(t *testing.T) {
	gql := `enum Hello { WORLD, HERE }`

	expected := graphql.NewEnum(graphql.EnumConfig{
		Name: "Hello",
		Values: graphql.EnumValueConfigMap{
			"WORLD": &graphql.EnumValueConfig{
				Value: 0,
			},
			"HERE": &graphql.EnumValueConfig{
				Value: 1,
			},
		},
	})

	ctx, _ := Generate(gql)
	enum := ctx.Enums("Hello")
	if !reflect.DeepEqual(enum, expected) {
		printFail(expected, enum, t)
	}
}

func TestSimpleFieldWithArg(t *testing.T) {
	gql := `
type Hello {
	world(flag: Boolean): String
}
	`

	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Hello",
		Fields: graphql.Fields{
			"world": &graphql.Field{
				Type: graphql.String,
				Args: graphql.FieldConfigArgument{
					"flag": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
				},
			},
		},
	})

	ctx, _ := Generate(gql)
	hello := ctx.Object("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestSimpleFieldWithArgDefaultValue(t *testing.T) {
	gql := `
type Hello {
	world(flag: Boolean! = true): String
}
	`

	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Hello",
		Fields: graphql.Fields{
			"world": &graphql.Field{
				Type: graphql.String,
				Args: graphql.FieldConfigArgument{
					"flag": &graphql.ArgumentConfig{
						Type:         graphql.NewNonNull(graphql.Boolean),
						DefaultValue: true,
					},
				},
			},
		},
	})

	ctx, _ := Generate(gql)
	hello := ctx.Object("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestSimpleFieldWithListArg(t *testing.T) {
	gql := `
type Hello {
	world(things: [String]): String
}
	`

	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Hello",
		Fields: graphql.Fields{
			"world": &graphql.Field{
				Type: graphql.String,
				Args: graphql.FieldConfigArgument{
					"things": &graphql.ArgumentConfig{
						Type: graphql.NewList(graphql.String),
					},
				},
			},
		},
	})

	ctx, _ := Generate(gql)
	hello := ctx.Object("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestSimpleFieldWithTwoArg(t *testing.T) {
	gql := `
type Hello {
	world(argOne: Boolean, argTwo: Int): String
}
	`

	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Hello",
		Fields: graphql.Fields{
			"world": &graphql.Field{
				Type: graphql.String,
				Args: graphql.FieldConfigArgument{
					"argOne": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
					"argTwo": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
			},
		},
	})

	ctx, _ := Generate(gql)
	hello := ctx.Object("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestSimpleUnion(t *testing.T) {
	gql := `
union Hello = World
type World {}`

	world := graphql.NewObject(graphql.ObjectConfig{
		Name: "World",
	})
	expected := graphql.NewUnion(graphql.UnionConfig{
		Name:  "Hello",
		Types: []*graphql.Object{world},
	})

	ctx, _ := Generate(gql)
	hello := ctx.Union("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestMultiUnion(t *testing.T) {
	gql := `
type Wor {}
type ld {}
union Hello = Wor | ld`

	wor := graphql.NewObject(graphql.ObjectConfig{
		Name: "Wor",
	})
	ld := graphql.NewObject(graphql.ObjectConfig{
		Name: "ld",
	})
	expected := graphql.NewUnion(graphql.UnionConfig{
		Name:  "Hello",
		Types: []*graphql.Object{wor, ld},
	})

	ctx, _ := Generate(gql)
	hello := ctx.Union("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestScalar(t *testing.T) {
	gql := `scalar Hello`

	expected := graphql.NewScalar(graphql.ScalarConfig{
		Name: "Hello",
	})

	ctx, _ := Generate(gql)
	hello := ctx.Scalar("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestSimpleInputObject(t *testing.T) {
	gql := `
input Hello {
	world: String
}`

	expected := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "Hello",
		Fields: graphql.InputObjectConfigFieldMap{
			"world": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	})

	ctx, _ := Generate(gql)
	hello := ctx.InputObject("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}

func TestSimpleExtendType(t *testing.T) {
	gql := `
type Hello {
  test: Boolean
}
extend type Hello {
  world: String
}`
	expected := graphql.NewObject(graphql.ObjectConfig{
		Name: "Hello",
		Fields: graphql.Fields{
			"test": &graphql.Field{
				Type: graphql.Boolean,
			},
			"world": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	ctx, _ := Generate(gql)
	hello := ctx.Object("Hello")
	if !reflect.DeepEqual(hello, expected) {
		printFail(expected, hello, t)
	}
}
