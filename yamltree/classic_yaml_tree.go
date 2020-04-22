package yamltree

import (
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

// ClassicDocumentNode is a wrapper type for a YAML document
type ClassicDocumentNode struct {
	doc interface{}
}

func (n ClassicDocumentNode) Kind() yaml.Kind {
	return yaml.DocumentNode
}

func (n ClassicDocumentNode) Get() YamlNode {
	node, err := toClassicNode(n.doc)
	if err != nil {
		// TODO: do something sensible here, like propagating an error
		return nil
	}
	return node
}

// ClassicSequenceNode is a wrapper type for a YAML sequence
type ClassicSequenceNode struct {
	sequence []interface{}
}

func (n ClassicSequenceNode) Kind() yaml.Kind {
	return yaml.SequenceNode
}

func (n ClassicSequenceNode) Len() int {
	return len(n.sequence)
}

func (n ClassicSequenceNode) GetAt(idx int) YamlNode {
	node, err := toClassicNode(n.sequence[idx])
	if err != nil {
		// TODO: do something sensible here, like propagating an error
		return nil
	}
	return node
}

func (n ClassicSequenceNode) InsertAt(node YamlNode, idx int) YamlSequence {
	obj, err := fromClassicNode(node)
	if err != nil {
		// TODO: do something sensible here, like propagating an error
		return &n
	}
	newAry := []interface{}{}
	newAry = append(newAry, n.sequence[:idx]...) // not inclusive
	newAry = append(newAry, obj)
	newAry = append(newAry, n.sequence[idx:]...) // inclusive
	n.sequence = newAry
	return &n
}

func (n ClassicSequenceNode) ReplaceAt(node YamlNode, idx int) YamlSequence {
	obj, err := fromClassicNode(node)
	if err != nil {
		// TODO: do something sensible here, like propagating an error
		fmt.Fprintf(os.Stderr, "ERROR: can't update index [%d] at node [%#v]: %s", idx, node, err.Error())
		return &n
	}
	n.sequence[idx] = obj
	return &n
}

func (n ClassicSequenceNode) Append(node YamlNode) YamlSequence {
	obj, err := fromClassicNode(node)
	if err != nil {
		// TODO: do something sensible here, like propagating an error
		return &n
	}
	n.sequence = append(n.sequence, obj)
	return &n
}

func (n ClassicSequenceNode) RemoveAt(idx int) YamlSequence {
	newAry := []interface{}{}
	newAry = append(newAry, n.sequence[:idx]...)
	newAry = append(newAry, n.sequence[idx+1:]...)
	n.sequence = newAry
	return &n
}

func (n ClassicSequenceNode) Each(visit func(YamlNode, int)) {
	for index, item := range n.sequence {
		node, err := toClassicNode(item)
		if err != nil {
			// TODO: do something sensible here, like propagating an error
			continue
		}
		visit(node, index)
	}
}

// ClassicMappingNode is a wrapper type for a YAML mapping
type ClassicMappingNode struct {
	mapping map[interface{}]interface{}
}

func (n ClassicMappingNode) Kind() yaml.Kind {
	return yaml.MappingNode
}

func (n ClassicMappingNode) Len() int {
	return len(n.mapping)
}

func (n ClassicMappingNode) Get(key string) (YamlNode, bool) {
	elem, found := n.mapping[key]
	if !found {
		return nil, false
	}
	node, err := toClassicNode(elem)
	if err != nil {
		return nil, false
	}
	return node, true
}

func (n ClassicMappingNode) Matches(key string, val string) bool {
	return n.mapping[key] == val
}

func (n ClassicMappingNode) Replace(key string, node YamlNode) {
	obj, err := fromClassicNode(node)
	if err != nil {
		// TODO: do something sensible here, like propagating an error
		fmt.Fprintf(os.Stderr, "ERROR: can't replace key [%s] at node [%#v]: %s", key, node, err.Error())
		return
	}
	n.mapping[key] = obj
}

func (n ClassicMappingNode) Remove(key string) {
	delete(n.mapping, key)
}

func (n ClassicMappingNode) EachKeys(visit func(string)) {
	for key := range n.mapping {
		keyStr, ok := key.(string)
		if !ok {
			continue
		}
		visit(keyStr)
	}
}

// ClassicScalarNode is a wrapper type for a YAML scalar
type ClassicScalarNode struct {
	scalar interface{}
}

func (n ClassicScalarNode) Kind() yaml.Kind {
	return yaml.ScalarNode
}

func (n ClassicScalarNode) String() string {
	return n.scalar.(string)
}

func fromClassicNode(n YamlNode) (interface{}, error) {
	if n == nil {
		return nil, nil
	}
	switch typedNode := n.(type) {
	case *ClassicDocumentNode:
		return typedNode.doc, nil
	case *ClassicSequenceNode:
		return typedNode.sequence, nil
	case *ClassicMappingNode:
		return typedNode.mapping, nil
	case *ClassicScalarNode:
		return typedNode.scalar, nil
	default:
		return nil, fmt.Errorf("unexpected node type %T", typedNode)
	}
}

func toClassicNode(n interface{}) (YamlNode, error) {
	if n == nil {
		return nil, nil
	}
	val := reflect.ValueOf(n)
	if !val.IsValid() {
		return nil, fmt.Errorf("Invalid value")
	}
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Slice:
		return &ClassicSequenceNode{
			sequence: val.Interface().([]interface{}),
		}, nil
	case reflect.Map:
		return &ClassicMappingNode{
			mapping: val.Interface().(map[interface{}]interface{}),
		}, nil
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return &ClassicScalarNode{
			scalar: val.Interface(),
		}, nil
	default:
		return nil, fmt.Errorf("unexpected node kind [%d]", val.Kind())
	}
}
