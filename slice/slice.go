package toolkit_slice

import (
	"reflect"
	"strings"
)


func In(source interface{}, target interface{}) bool {

	v := reflect.ValueOf(target)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		var array = target.([]interface{})
		for i := 0; i < len(array); i++ {
			if array[i] == source {
				return true
			}
		}
	case reflect.String:
		if _, ok := source.(string); !ok {
			return false
		}
		return strings.Contains(target.(string), source.(string))
	default:
		return false
	}

	return false
}