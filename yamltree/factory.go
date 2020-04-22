package yamltree

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func CreateYamlDocumentV2(input []byte) (YamlNode, error) {
	var tree interface{}

	err := yaml.Unmarshal(input, &tree)
	if err != nil {
		return nil, err
	}

	yamlDocument := newClassicDocumentNode(tree)

	return yamlDocument, nil
}

func CreateYamlNodeV2(tree interface{}) YamlNode {
	if tree == nil {
		return nil
	}
	node, err := toClassicNode(tree)
	if err != nil {
		// TODO: do something sensible here, like propagating an error
		fmt.Fprintf(os.Stderr, "ERROR: can't create YAML node out of [%#v]: %s", tree, err.Error())
		return nil
	}
	return node
}

func newClassicDocumentNode(tree interface{}) YamlDocument {
	return &ClassicDocumentNode{
		doc: tree,
	}
}

func CreateYamlSequenceV2() YamlSequence {
	return &ClassicSequenceNode{
		sequence: []interface{}{},
	}
}

func CreateStockYamlSequenceV2(seq []interface{}) YamlSequence {
	return &ClassicSequenceNode{
		sequence: seq,
	}
}

func CreateYamlMappingV2() YamlMapping {
	return &ClassicMappingNode{
		mapping: map[interface{}]interface{}{},
	}
}

func CreateSingleKeyYamlMappingV2(key string, value string) YamlMapping {
	return &ClassicMappingNode{
		mapping: map[interface{}]interface{}{key: value},
	}
}
