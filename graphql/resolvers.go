package graphql

import (
	"reflect"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
)

//
// ScalarResolver represents a collection of methods whose products represent
// the input and response values of a scalar type.
//
// == Example input SDL
//
//   """
//   Timestamps are great.
//   """
//   scalar Timestamp
//
// == Example implementation
//
//   // MyTimestampResolver implements ScalarResolver interface
//   type MyTimestampResolver struct {
//     defaultTZ *time.Location
//     logger    logrus.LogEntry
//   }
//
//   // Serialize serializes given date into RFC 943 compatible string.
//   func (r *MyTimestampResolver) Serialize(val interface{}) interface{} {
//     // ... implementation details ...
//   }
//
//   // ParseValue takes given value and coerces it into an instance of Time.
//   func (r *MyTimestampResolver) ParseValue(val interface{}) interface{} {
//     // ... implementation details ...
//     // eg. if val is an int use time.At(), if string time.Parse(), etc.
//   }
//
//   // ParseValue takes given value and coerces it into an instance of Time.
//   func (r *MyTimestampResolver) ParseValue(val ast.Value) interface{} {
//     // ... implementation details ...
//     //
//     // eg.
//     //
//     // if string value return value,
//     // if IntValue Atoi and return value,
//     // etc.
//   }`
//
type ScalarResolver interface {
	// Serialize an internal value to include in a response.
	Serialize(interface{}) interface{}

	// ParseValue parses an externally provided value to use as an input.
	ParseValue(interface{}) interface{}

	// ParseLiteral parses an externally provided literal value to use as an input.
	ParseLiteral(ast.Value) interface{}
}

// DefaultResolver uses reflection to attempt to resolve the result of a given
// field.
func DefaultResolver(source interface{}, fieldName string) (interface{}, error) {
	// Heavily borrows from: https://github.com/graphql-go/graphql/blob/9b68c99d07d901738c15564ec1a0f57d07d884a7/executor.go#L823-L881
	sourceVal := reflect.ValueOf(source)
	if sourceVal.IsValid() && sourceVal.Type().Kind() == reflect.Ptr {
		sourceVal = sourceVal.Elem()
	}
	if !sourceVal.IsValid() {
		return nil, nil
	}

	// Struct
	if sourceVal.Type().Kind() == reflect.Struct {
		for i := 0; i < sourceVal.NumField(); i++ {
			valueField := sourceVal.Field(i)
			typeField := sourceVal.Type().Field(i)
			// try matching the field name first
			if typeField.Name == fieldName {
				return valueField.Interface(), nil
			}
			tag := typeField.Tag
			checkTag := func(tagName string) bool {
				t := tag.Get(tagName)
				tOptions := strings.Split(t, ",")
				if len(tOptions) == 0 {
					return false
				}
				if tOptions[0] != fieldName {
					return false
				}
				return true
			}
			if checkTag("json") || checkTag("graphql") {
				return valueField.Interface(), nil
			}
			continue
		}
		return nil, nil
	}

	// map[string]interface
	if sourceMap, ok := source.(map[string]interface{}); ok {
		property := sourceMap[fieldName]
		val := reflect.ValueOf(property)
		if val.IsValid() && val.Type().Kind() == reflect.Func {
			// try type casting the func to the most basic func signature
			// for more complex signatures, user have to define ResolveFn
			if propertyFn, ok := property.(func() interface{}); ok {
				return propertyFn(), nil
			}
		}
		return property, nil
	}

	// last resort, return nil
	return nil, nil
}
