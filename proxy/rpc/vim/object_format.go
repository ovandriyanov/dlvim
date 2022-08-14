package vim

import (
	"encoding/json"
	"reflect"
	"strings"

	dlvapi "github.com/go-delve/delve/service/api"
)

// Converts a variable to an object ready to be marshaled into a JSON
func toObject(variable *dlvapi.Variable) (object interface{}) {
	switch variable.Kind {
	case reflect.Invalid:
		return nil
	case reflect.Pointer, reflect.Interface:
		if len(variable.Children) == 0 {
			object = nil
		} else {
			object = toObject(&variable.Children[0])
		}
	case reflect.Array, reflect.Slice:
		list := make([]interface{}, 0, len(variable.Children))
		for _, child := range variable.Children {
			list = append(list, toObject(&child))
		}
		object = list
	default:
		if len(variable.Children) == 0 {
			return variable.Value
		}
		dict := make(map[string]interface{})
		for _, child := range variable.Children {
			dict[child.Name] = toObject(&child)
		}
		object = dict
	}
	return object
}

func formatVariable(variable *dlvapi.Variable) []string {
	object := toObject(variable)
	marshaled, _ := json.MarshalIndent(object, "", "  ")
	return strings.Split(string(marshaled), "\n")
}
