package yamltree

import (
	"gopkg.in/yaml.v3"
)

type YamlNode interface {
	Kind() yaml.Kind
}

type YamlDocument interface {
	YamlNode
	Get() YamlNode
}

type YamlSequence interface {
	YamlNode
	Len() int
	GetAt(int) YamlNode
	InsertAt(YamlNode, int) YamlSequence  // TODO: reverse arguments order
	ReplaceAt(YamlNode, int) YamlSequence // TODO: reverse arguments order
	Append(YamlNode) YamlSequence
	RemoveAt(int) YamlSequence
	Each(func(YamlNode, int)) // TODO: reverse arguments order
}

type YamlMapping interface {
	YamlNode
	Len() int
	Get(string) (YamlNode, bool)
	Matches(string, string) bool
	Replace(string, YamlNode)
	Remove(string)
	EachKeys(func(string))
}

type YamlScalar interface {
	YamlNode
	String() string
}
