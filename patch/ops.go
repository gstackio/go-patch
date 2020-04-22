package patch

import (
	"github.com/cppforlife/go-patch/yamltree"
)

type Ops []Op

type Op interface {
	Apply(yamltree.YamlNode) (yamltree.YamlNode, error)
}

// Ensure basic operations implement Op
var _ Op = Ops{}
var _ Op = ReplaceOp{}
var _ Op = RemoveOp{}
var _ Op = FindOp{}
var _ Op = DescriptiveOp{}
var _ Op = ErrOp{}

func (ops Ops) Apply(doc yamltree.YamlNode) (yamltree.YamlNode, error) {
	var err error

	for _, op := range ops {
		doc, err = op.Apply(doc)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}
