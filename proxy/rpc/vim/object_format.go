package vim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	dlvapi "github.com/go-delve/delve/service/api"
)

type structField struct {
	name  string
	value interface{}
}

type goStruct struct {
	typeName string
	fields   []structField
}

// Converts a variable to an object ready to be marshaled into a JSON
func toValue(variable *dlvapi.Variable) (object interface{}) {
	switch variable.Kind {
	case reflect.Invalid:
		return nil
	case reflect.Pointer, reflect.Interface:
		if len(variable.Children) == 0 {
			object = nil
		} else {
			object = toValue(&variable.Children[0])
		}
	case reflect.Array, reflect.Slice:
		list := make([]interface{}, 0, len(variable.Children))
		for _, child := range variable.Children {
			list = append(list, toValue(&child))
		}
		object = list
	default:
		if len(variable.Children) == 0 {
			return variable.Value
		}
		structValue := goStruct{
			typeName: variable.Type,
			fields:   make([]structField, 0, len(variable.Children)),
		}
		for _, child := range variable.Children {
			structValue.fields = append(structValue.fields, structField{name: child.Name, value: toValue(&child)})
		}
		object = structValue
	}
	return object
}

func formatVariable(variable *dlvapi.Variable) []string {
	value := toValue(variable)
	marshaled := marshalValue(value, 0)
	return strings.Split(string(marshaled), "\n")
}

func marshalValue(value interface{}, nestingLevel int) (marshaled []byte) {
	switch typedValue := value.(type) {
	case goStruct:
		marshaled = marshalGoStruct(typedValue, nestingLevel)
	case []interface{}:
		marshaled = marshalSlice(typedValue, nestingLevel)
	default:
		marshaled, _ = json.MarshalIndent(value, indent(nestingLevel), "  ")
	}
	return marshaled
}

func marshalSlice(slice []interface{}, nestingLevel int) (marshaled []byte) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(fmt.Sprintf("[\n"))
	for i, value := range slice {
		buf.WriteString(fmt.Sprintf("%s  %s", indent(nestingLevel), string(marshalValue(value, nestingLevel+1))))
		if i != len(slice)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintf("%s]", indent(nestingLevel)))
	return buf.Bytes()
}

func marshalGoStruct(structValue goStruct, nestingLevel int) []byte {
	buf := bytes.NewBuffer(nil)

	maybeComma := ""
	if len(structValue.fields) > 0 {
		maybeComma = ","
	}

	buf.WriteString(fmt.Sprintf("{ \"#type\": %q%s\n", structValue.typeName, maybeComma))
	for i, field := range structValue.fields {
		marshaledField := marshalValue(field.value, nestingLevel+1)
		maybeComma = ","
		if i == len(structValue.fields)-1 {
			maybeComma = ""
		}
		buf.WriteString(fmt.Sprintf("%s  %q: %s%s\n", indent(nestingLevel), field.name, string(marshaledField), maybeComma))
	}
	buf.WriteString(fmt.Sprintf("%s}", indent(nestingLevel)))
	return buf.Bytes()
}

func indent(nestingLevel int) (result string) {
	for i := 0; i < nestingLevel; i++ {
		result = result + "  "
	}
	return result
}
