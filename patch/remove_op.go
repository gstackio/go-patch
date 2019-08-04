package patch

import (
	"fmt"
	"reflect"
)

type RemoveOp struct {
	Path Pointer
}

func (op RemoveOp) Apply(doc interface{}) (interface{}, error) {
	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return nil, fmt.Errorf("Cannot remove entire document")
	}

	obj := doc
	prevUpdate := func(newObj interface{}) { doc = newObj }

	for i, token := range tokens[1:] {
		isLast := i == len(tokens)-2
		currPath := NewPointer(tokens[:i+2])

		switch typedToken := token.(type) {
		case IndexToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Slice {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: ptr, Path: currPath}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				newAry := reflect.ValueOf([]interface{}{})
				newAry = reflect.AppendSlice(newAry, ptr.Slice(0, idx))           // not inclusive
				newAry = reflect.AppendSlice(newAry, ptr.Slice(idx+1, ptr.Len())) // inclusive

				prevUpdate(newAry.Interface())
			} else {
				obj = ptr.Index(idx).Interface()
				prevUpdate = func(newObj interface{}) {
					ptr.Index(idx).Set(reflect.ValueOf(newObj))
				}
			}

		case MatchingIndexToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Slice {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			idxs := findMapIndices(ptr, typedToken.Key, typedToken.Value)

			if typedToken.Optional && len(idxs) == 0 {
				return doc, nil
			}

			if len(idxs) != 1 {
				return nil, OpMultipleMatchingIndexErr{currPath, idxs}
			}

			idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: ptr, Path: currPath}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				newAry := reflect.ValueOf([]interface{}{})
				newAry = reflect.AppendSlice(newAry, ptr.Slice(0, idx))           // not inclusive
				newAry = reflect.AppendSlice(newAry, ptr.Slice(idx+1, ptr.Len())) // inclusive
				prevUpdate(newAry.Interface())
			} else {
				obj = ptr.Index(idx).Interface()
				// no need to change prevUpdate since matching item can only be a map
			}

		case KeyToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Map {
				return nil, NewOpMapMismatchTypeErr(currPath, obj)
			}

			if mapValue := ptr.MapIndex(reflect.ValueOf(typedToken.Key)); !mapValue.IsValid() {
				if typedToken.Optional {
					return doc, nil
				}
				return nil, OpMissingMapKeyErr{typedToken.Key, currPath, ptr}
			} else {
				obj = mapValue.Interface()
			}

			if isLast {
				ptr.SetMapIndex(reflect.ValueOf(typedToken.Key), reflect.Value{})
			} else {
				prevUpdate = func(newObj interface{}) {
					ptr.SetMapIndex(reflect.ValueOf(typedToken.Key), reflect.ValueOf(newObj))
				}
			}

		default:
			return nil, OpUnexpectedTokenErr{token, currPath}
		}
	}

	return doc, nil
}
