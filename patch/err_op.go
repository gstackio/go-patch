package patch

import (
	"github.com/cppforlife/go-patch/yamltree"
)

type ErrOp struct {
	Err error
}

func (op ErrOp) Apply(_ yamltree.YamlNode) (yamltree.YamlNode, error) {
	return nil, op.Err
}
