package patch

import (
	"fmt"
	"reflect"
)

type ArrayIndex struct {
	Index     int
	Modifiers []Modifier
	Array     reflect.Value
	Path      Pointer
}

func (i ArrayIndex) Concrete() (int, error) {
	result := i.Index

	for _, modifier := range i.Modifiers {
		switch modifier.(type) {
		case PrevModifier:
			result -= 1
		case NextModifier:
			result += 1
		default:
			return 0, fmt.Errorf("Expected to find one of the following modifiers: 'prev', 'next', but found modifier '%T'", modifier)
		}
	}

	if result >= i.Array.Len() || (-result)-1 >= i.Array.Len() {
		return 0, OpMissingIndexErr{result, i.Array, i.Path}
	}

	if result < 0 {
		result = i.Array.Len() + result
	}

	return result, nil
}
