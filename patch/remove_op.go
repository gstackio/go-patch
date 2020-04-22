package patch

import (
	"fmt"

	"github.com/cppforlife/go-patch/yamltree"
)

type RemoveOp struct {
	Path Pointer
}

func (op RemoveOp) Apply(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return nil, fmt.Errorf("Cannot remove entire document")
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

			idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				prevUpdate(typedObj.RemoveAt(idx))
			} else {
				obj = typedObj.GetAt(idx)
				prevUpdate = func(newObj yamltree.YamlNode) { typedObj.ReplaceAt(newObj, idx) }
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
				return doc, nil
			}

			if len(idxs) != 1 {
				return nil, OpMultipleMatchingIndexErr{currPath, idxs}
			}

			idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				prevUpdate(typedObj.RemoveAt(idx))
			} else {
				obj = typedObj.GetAt(idx)
				prevUpdate = func(newObj yamltree.YamlNode) { typedObj.ReplaceAt(newObj, idx) }
			}

		case KeyToken:
			typedObj, ok := obj.(yamltree.YamlMapping)
			if !ok {
				return nil, NewOpMapMismatchTypeErr(currPath, obj)
			}

			var found bool

			obj, found = typedObj.Get(typedToken.Key)
			if !found {
				if typedToken.Optional {
					return doc, nil
				}

				return nil, OpMissingMapKeyErr{typedToken.Key, currPath, typedObj}
			}

			if isLast {
				typedObj.Remove(typedToken.Key)
			} else {
				key := typedToken.Key
				prevUpdate = func(newObj yamltree.YamlNode) { typedObj.Replace(key, newObj) }
			}

		default:
			return nil, OpUnexpectedTokenErr{token, currPath}
		}
	}

	return doc, nil
}
