package patch

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

type ReplaceOp struct {
	Path  Pointer
	Value interface{} // will be cloned using yaml library
}

func (op ReplaceOp) Apply(doc interface{}) (interface{}, error) {
	// Ensure that value is not modified by future operations
	clonedValue, err := op.cloneValue(op.Value)
	if err != nil {
		return nil, fmt.Errorf("ReplaceOp cloning value: %s", err)
	}

	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return clonedValue, nil
	}

	obj := doc
	prevUpdate := func(newObj interface{}) {
		doc = newObj
	}

	for i, token := range tokens[1:] {
		isLast := i == len(tokens)-2
		currPath := NewPointer(tokens[:i+2])

		switch typedToken := token.(type) {
		case IndexToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Slice {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			if isLast {
				idx, err := ArrayInsertion{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: ptr, Path: currPath}.Concrete()
				if err != nil {
					return nil, err
				}

				prevUpdate(idx.Update(ptr, clonedValue))
			} else {
				idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: ptr, Path: currPath}.Concrete()
				if err != nil {
					return nil, err
				}

				obj = ptr.Index(idx).Interface()
				prevUpdate = func(newObj interface{}) { ptr.Index(idx).Set(reflect.ValueOf(newObj)) }
			}

		case AfterLastIndexToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Slice {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			if isLast {
				prevUpdate(reflect.Append(ptr, reflect.ValueOf(clonedValue)).Interface())
			} else {
				return nil, fmt.Errorf("Expected after last index token to be last in path '%s'", op.Path)
			}

		case MatchingIndexToken:
			ptr := reflect.ValueOf(obj)
			if ptr.Kind() != reflect.Slice {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			idxs := findMapIndices(ptr, typedToken.Key, typedToken.Value)

			if typedToken.Optional && len(idxs) == 0 {
				if isLast {
					prevUpdate(reflect.Append(ptr, reflect.ValueOf(clonedValue)).Interface())
				} else {
					obj = map[interface{}]interface{}{typedToken.Key: typedToken.Value}
					prevUpdate(reflect.Append(ptr, reflect.ValueOf(obj)).Interface())
					// no need to change prevUpdate since matching item can only be a map
				}
			} else {
				if len(idxs) != 1 {
					return nil, OpMultipleMatchingIndexErr{currPath, idxs}
				}

				if isLast && len(idxs) == 1 {
					idx, err := ArrayInsertion{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: ptr, Path: currPath}.Concrete()
					if err != nil {
						return nil, err
					}

					prevUpdate(idx.Update(ptr, clonedValue))
				} else if len(idxs) == 1 {
					idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: ptr, Path: currPath}.Concrete()
					if err != nil {
						return nil, err
					}

					obj = ptr.Index(idx).Interface()
					// no need to change prevUpdate since matching item can only be a map
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

			setValue := func(value interface{}) {
				v := reflect.ValueOf(value)

				if !v.IsValid() && ptr.Type().Elem().Kind() == reflect.Interface {
					v = reflect.Zero(ptr.Type().Elem())
				}

				ptr.SetMapIndex(reflect.ValueOf(typedToken.Key), v)
			}

			if isLast {
				setValue(clonedValue)
			} else {
				prevUpdate = func(newObj interface{}) { setValue(newObj) }

				if !found {
					// Determine what type of value to create based on next token
					switch tokens[i+2].(type) {
					case AfterLastIndexToken:
						obj = []interface{}{}
					case MatchingIndexToken:
						obj = []interface{}{}
					case KeyToken:
						obj = map[interface{}]interface{}{}
					default:
						errMsg := "Expected to find key, matching index or after last index token at path '%s'"
						return nil, fmt.Errorf(errMsg, NewPointer(tokens[:i+3]))
					}

					setValue(obj)
				}
			}

		default:
			return nil, OpUnexpectedTokenErr{token, currPath}
		}
	}

	return doc, nil
}

func (ReplaceOp) cloneValue(in interface{}) (out interface{}, err error) {
	defer func() {
		if recoverVal := recover(); recoverVal != nil {
			err = fmt.Errorf("Recovered: %s", recoverVal)
		}
	}()

	bytes, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
