package patch

import (
	"fmt"
	"reflect"

	"github.com/cppforlife/go-patch/yamltree"
)

type ReplaceOp struct {
	Path  Pointer
	Value interface{} // will be cloned
}

func (op ReplaceOp) Apply(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
	// Ensure that value is not modified by future operations
	clonedTree, err := cloneTree(op.Value)
	if err != nil {
		return nil, fmt.Errorf("ReplaceOp cloning value: %s", err)
	}
	clonedValue := yamltree.CreateYamlNodeV2(clonedTree)

	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return clonedValue, nil
	}

	obj := doc
	prevUpdate := func(newObj yamltree.YamlNode) { doc = newObj }

	for i, token := range tokens[1:] {
		isLast := i == len(tokens)-2
		currPath := NewPointer(tokens[:i+2])

		switch typedToken := token.(type) {
		case IndexToken:
			typedObj, ok := obj.(yamltree.YamlSequence)
			if !ok {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			if isLast {
				idx, err := ArrayInsertion{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
				if err != nil {
					return nil, err
				}

				prevUpdate(idx.Update(typedObj, clonedValue))
			} else {
				idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
				if err != nil {
					return nil, err
				}

				obj = typedObj.GetAt(idx)
				prevUpdate = func(newObj yamltree.YamlNode) { typedObj.ReplaceAt(newObj, idx) }
			}

		case AfterLastIndexToken:
			typedObj, ok := obj.(yamltree.YamlSequence)
			if !ok {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			if isLast {
				prevUpdate(typedObj.Append(clonedValue))
			} else {
				return nil, fmt.Errorf("Expected after last index token to be last in path '%s'", op.Path)
			}

		case MatchingIndexToken:
			typedObj, ok := obj.(yamltree.YamlSequence)
			if !ok {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			var idxs []int

			typedObj.Each(func(item yamltree.YamlNode, itemIdx int) {
				typedItem, ok := item.(yamltree.YamlMapping)
				if ok {
					if typedItem.Matches(typedToken.Key, typedToken.Value) {
						idxs = append(idxs, itemIdx)
					}
				}
			})

			if typedToken.Optional && len(idxs) == 0 {
				if isLast {
					prevUpdate(typedObj.Append(clonedValue))
				} else {
					obj = yamltree.CreateSingleKeyYamlMappingV2(typedToken.Key, typedToken.Value)
					prevUpdate(typedObj.Append(obj))
					idx := typedObj.Len() - 1
					prevUpdate = func(newObj yamltree.YamlNode) { typedObj.ReplaceAt(newObj, idx) }
				}
			} else {
				if len(idxs) != 1 {
					return nil, OpMultipleMatchingIndexErr{currPath, idxs}
				}

				if isLast {
					idx, err := ArrayInsertion{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
					if err != nil {
						return nil, err
					}

					prevUpdate(idx.Update(typedObj, clonedValue))
				} else {
					idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
					if err != nil {
						return nil, err
					}

					obj = typedObj.GetAt(idx)
					prevUpdate = func(newObj yamltree.YamlNode) { typedObj.ReplaceAt(newObj, idx) }
				}
			}

		case KeyToken:
			typedObj, ok := obj.(yamltree.YamlMapping)
			if !ok {
				return nil, NewOpMapMismatchTypeErr(currPath, obj)
			}

			var found bool

			obj, found = typedObj.Get(typedToken.Key)
			if !found && !typedToken.Optional {
				return nil, OpMissingMapKeyErr{typedToken.Key, currPath, typedObj}
			}

			if isLast {
				typedObj.Replace(typedToken.Key, clonedValue)
			} else {
				prevUpdate = func(newObj yamltree.YamlNode) { typedObj.Replace(typedToken.Key, newObj) }

				if !found {
					// Determine what type of value to create based on next token
					switch tokens[i+2].(type) {
					case AfterLastIndexToken:
						obj = yamltree.CreateYamlSequenceV2()
					case MatchingIndexToken:
						obj = yamltree.CreateYamlSequenceV2()
					case KeyToken:
						obj = yamltree.CreateYamlMappingV2()
					default:
						errMsg := "Expected to find key, matching index or after last index token at path '%s'"
						return nil, fmt.Errorf(errMsg, NewPointer(tokens[:i+3]))
					}

					typedObj.Replace(typedToken.Key, obj)
				}
			}

		default:
			return nil, OpUnexpectedTokenErr{token, currPath}
		}
	}

	return doc, nil
}

func cloneTree(in interface{}) (out interface{}, err error) {
	if in == nil {
		return nil, nil
	}
	val := reflect.ValueOf(in)
	if !val.IsValid() {
		return nil, fmt.Errorf("can't clone invalid value: %#v", out)
	}
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Slice:
		array := val.Interface().([]interface{})
		newArray := make([]interface{}, len(array))
		for idx := range array {
			node, err := cloneTree(array[idx])
			if err != nil {
				return nil, err
			}
			newArray[idx] = node
		}
		out = newArray
	case reflect.Map:
		newMap := map[interface{}]interface{}{}
		for k, v := range val.Interface().(map[interface{}]interface{}) {
			node, err := cloneTree(v)
			if err != nil {
				return nil, err
			}
			newMap[k] = node
		}
		out = newMap
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		out = in
	default:
		return nil, fmt.Errorf("unexpected node kind [%d]", val.Kind())
	}
	return out, nil
}
