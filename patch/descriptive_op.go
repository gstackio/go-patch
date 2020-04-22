package patch

import (
	"fmt"

	"github.com/cppforlife/go-patch/yamltree"
)

type DescriptiveOp struct {
	Op       Op
	ErrorMsg string
}

func (op DescriptiveOp) Apply(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
	doc, err := op.Op.Apply(doc)
	if err != nil {
		return nil, fmt.Errorf("Error '%s': %s", op.ErrorMsg, err.Error())
	}
	return doc, nil
}
