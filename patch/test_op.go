package patch

import (
	"fmt"
	"reflect"

	"github.com/cppforlife/go-patch/yamltree"
)

type TestOp struct {
	Path   Pointer
	Value  interface{}
	Absent bool
}

func (op TestOp) Apply(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
	if op.Absent {
		return op.checkAbsence(doc)
	}
	return op.checkValue(doc)
}

func (op TestOp) checkAbsence(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
	_, err := FindOp{Path: op.Path}.Apply(doc)
	if err != nil {
		if typedErr, ok := err.(OpMissingIndexErr); ok {
			if typedErr.Path.String() == op.Path.String() {
				return doc, nil
			}
		}
		if typedErr, ok := err.(OpMissingMapKeyErr); ok {
			if typedErr.Path.String() == op.Path.String() {
				return doc, nil
			}
		}
		return nil, err
	}

	return nil, fmt.Errorf("Expected to not find '%s'", op.Path)
}

func (op TestOp) checkValue(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
	foundVal, err := FindOp{Path: op.Path}.Apply(doc)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(foundVal, yamltree.CreateYamlNodeV2(op.Value)) {
		return nil, fmt.Errorf("Found value does not match expected value")
	}

	// Return same input document
	return doc, nil
}
