package patch

import (
	"fmt"

	"github.com/cppforlife/go-patch/yamltree"
)

type ArrayInsertion struct {
	Index     int
	Modifiers []Modifier
	Array     yamltree.YamlSequence
	Path      Pointer
}

type ArrayInsertionIndex struct {
	number int
	insert bool
}

func (i ArrayInsertion) Concrete() (ArrayInsertionIndex, error) {
	var mods []Modifier

	before := false
	after := false

	for _, modifier := range i.Modifiers {
		if before {
			return ArrayInsertionIndex{}, fmt.Errorf(
				"Expected to not find any modifiers after 'before' modifier, but found modifier '%T'", modifier)
		}
		if after {
			return ArrayInsertionIndex{}, fmt.Errorf(
				"Expected to not find any modifiers after 'after' modifier, but found modifier '%T'", modifier)
		}

		switch modifier.(type) {
		case BeforeModifier:
			before = true
		case AfterModifier:
			after = true
		default:
			mods = append(mods, modifier)
		}
	}

	idx := ArrayIndex{Index: i.Index, Modifiers: mods, Array: i.Array, Path: i.Path}

	num, err := idx.Concrete()
	if err != nil {
		return ArrayInsertionIndex{}, err
	}

	if after && num != i.Array.Len() {
		num += 1
	}

	return ArrayInsertionIndex{num, before || after}, nil
}

func (i ArrayInsertionIndex) Update(array yamltree.YamlSequence, obj yamltree.YamlNode) yamltree.YamlSequence {
	if i.insert {
		return array.InsertAt(obj, i.number)
	}

	return array.ReplaceAt(obj, i.number)
}
