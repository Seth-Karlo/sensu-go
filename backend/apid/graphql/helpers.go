package graphqlschema

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

//
// AliasResolver makes it quick and easy to create a resolver that reflects a
// different field on a given resource.
//
// Usage:
//
// "legs": &graphql.Field{
//   Name:    "number of legs the owner's cat has",
//   Type:    graphql.Int,
//   Resolve: AliasResolver("myCat", "numberOfLegs"),
// },
//
func AliasResolver(fNames ...string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		fVal := reflect.ValueOf(p.Source)
		for _, fName := range fNames {
			fVal = reflect.Indirect(fVal)
			fVal = fVal.FieldByName(fName)
		}
		return fVal.Interface(), nil
	}
}

//
// NewInputFromObjectField pulls in fields from a GraphQL type and returns
// an inputfield config.
//
// Usage:
//
// relay.MutationWithClientMutationID(relay.MutationConfig{
//   Name: "UpdateMyCat",
//   InputFields: graphql.InputObjectConfigFieldMap{
//     "name": AliasForInputField(CatType, "name", nil),
//     "paws": AliasForInputField(CatType, "pawsNum", 4),
//     "deletable": &graphql.InputObjectFieldConfig{
//       Type: graphql.NewNonNull(graphql.Boolean),
//       Description: "Whether or not the cat is deletable",
//       DefaultValue: true,
//     },
//   }),
//   ...
// })
//
func NewInputFromObjectField(obj *graphql.Object, fieldName string, defaultValue interface{}) *graphql.InputObjectFieldConfig {
	for _, field := range obj.Fields() {
		if field.Name != fieldName {
			continue
		}

		return &graphql.InputObjectFieldConfig{
			Type:         field.Type,
			Description:  field.Description,
			DefaultValue: defaultValue,
		}
	}

	logger.
		WithField("name", fieldName).
		WithField("object", obj).
		Panic("given field did not match any fields on type")

	return nil
}

// DecodeIDFromInputs takes inputs and a field name and will return the global id components for the
// input matching the field name
func DecodeIDFromInputs(inputs map[string]interface{}, fieldName string) (globalid.Components, error) {
	id, _ := inputs[fieldName].(string)
	return globalid.Decode(id)
}

// SetContextFromComponents takes a context and global id components, adds the environment and
// organization to the context, and returns the updated context
func SetContextFromComponents(ctx context.Context, c globalid.Components) context.Context {
	ctx = context.WithValue(ctx, types.EnvironmentKey, c.Environment)
	ctx = context.WithValue(ctx, types.OrganizationKey, c.Organization)
	return ctx
}

// IDQueryParamsFromComponents takes a global id components and returns QueryParams with the id key
// set to the components' primary id
func IDQueryParamsFromComponents(components globalid.Components) actions.QueryParams {
	return actions.QueryParams{
		"id": components.UniqueComponent(),
	}
}
