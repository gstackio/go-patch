package patch

import (
	"fmt"

	"github.com/cppforlife/go-patch/yamltree"
)

type FindOp struct {
	Path Pointer
}

func (op FindOp) Apply(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
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
			typedObj, ok := obj.(yamltree.YamlSequence)
			if !ok {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				return typedObj.GetAt(idx), nil
			} else {
				obj = typedObj.GetAt(idx)
			}

		case AfterLastIndexToken:
			errMsg := "Expected not to find after last index token in path '%s' (not supported in find operations)"
			return nil, fmt.Errorf(errMsg, op.Path)

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
				// todo /blah=foo?:after, modifiers
				obj = yamltree.CreateSingleKeyYamlMappingV2(typedToken.Key, typedToken.Value)

				if isLast {
					return obj, nil
				}
			} else {
				if len(idxs) != 1 {
					return nil, OpMultipleMatchingIndexErr{currPath, idxs}
				}

				idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
				if err != nil {
					return nil, err
				}

				if isLast {
					return typedObj.GetAt(idx), nil
				} else {
					obj = typedObj.GetAt(idx)
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
				val, _ := typedObj.Get(typedToken.Key)
				return val, nil
			} else {
				if !found {
					// Determine what type of value to create based on next token
					switch tokens[i+2].(type) {
					case MatchingIndexToken:
						obj = yamltree.CreateYamlSequenceV2()
					case KeyToken:
						obj = yamltree.CreateYamlMappingV2()
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
