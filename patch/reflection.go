package patch

import (
	"reflect"
)

func dereference(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		return v.Elem()
	default:
		return v
	}
}

func findMapIndices(sliceOfMaps reflect.Value, key, value interface{}) []int {
	var idxs []int

	for itemIdx := 0; itemIdx < sliceOfMaps.Len(); itemIdx++ {
		item := dereference(sliceOfMaps.Index(itemIdx))

		if item.Kind() != reflect.Map {
			continue
		}

		v := item.MapIndex(reflect.ValueOf(key))

		if !v.IsValid() {
			continue
		}


		if dereference(v).Interface() != value {
			continue
		}

		idxs = append(idxs, itemIdx)
	}

	return idxs
}