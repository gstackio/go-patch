package patch

import (
	"fmt"
	"reflect"
)

type FindOp struct {
	Path Pointer
}

func (op FindOp) Apply(doc interface{}) (interface{}, error) {
	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return doc, nil
	}

	obj := doc

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
				return ptr.Index(idx).Interface(), nil
			} else {
				obj = ptr.Index(idx).Interface()
			}

		case AfterLastIndexToken:
			errMsg := "Expected not to find after last index token in path '%s' (not supported in find operations)"
			return nil, fmt.Errorf(errMsg, op.Path)

		case MatchingIndexToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Slice {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			idxs := findMapIndices(ptr, typedToken.Key, typedToken.Value)


			if typedToken.Optional && len(idxs) == 0 {
				// todo /blah=foo?:after, modifiers
				obj = map[interface{}]interface{}{typedToken.Key: typedToken.Value}

				if isLast {
					return obj, nil
				}
			} else {
				if len(idxs) != 1 {
					return nil, OpMultipleMatchingIndexErr{currPath, idxs}
				}

				idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: ptr, Path: currPath}.Concrete()
				if err != nil {
					return nil, err
				}

				if isLast {
					return ptr.Index(idx).Elem().Interface(), nil
				} else {
					obj = ptr.Index(idx).Elem().Interface()
				}
			}

		case KeyToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Map {
				return nil, NewOpMapMismatchTypeErr(currPath, obj)
			}

			var found bool

			if mapValue := ptr.MapIndex(reflect.ValueOf(typedToken.Key)); mapValue.IsValid() {
				obj = mapValue.Interface()
				found = true
			} else {
				if !typedToken.Optional {
					return nil, OpMissingMapKeyErr{typedToken.Key, currPath, ptr}
				}

				found = false
				obj = nil
			}

			if isLast {
				if v := ptr.MapIndex(reflect.ValueOf(typedToken.Key)); v.IsValid() {
					return v.Interface(), nil
				}
				return nil, nil
			} else {
				if !found {
					// Determine what type of value to create based on next token
					switch tokens[i+2].(type) {
					case MatchingIndexToken:
						obj = []interface{}{}
					case KeyToken:
						obj = map[interface{}]interface{}{}
					default:
						errMsg := "Expected to find key or matching index token at path '%s'"
						return nil, fmt.Errorf(errMsg, NewPointer(tokens[:i+3]))
					}
				}
			}

		default:
			return nil, OpUnexpectedTokenErr{token, currPath}
		}
	}

	return doc, nil
}
