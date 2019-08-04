package patch

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type OpMismatchTypeErr struct {
	Type_ string
	Path  Pointer
	Obj   interface{}
}

func NewOpArrayMismatchTypeErr(path Pointer, obj interface{}) OpMismatchTypeErr {
	return OpMismatchTypeErr{"an array", path, obj}
}

func NewOpMapMismatchTypeErr(path Pointer, obj interface{}) OpMismatchTypeErr {
	return OpMismatchTypeErr{"a map", path, obj}
}

func (e OpMismatchTypeErr) Error() string {
	errMsg := "Expected to find %s at path '%s' but found '%T'"
	return fmt.Sprintf(errMsg, e.Type_, e.Path, e.Obj)
}

type OpMissingMapKeyErr struct {
	Key  string
	Path Pointer
	Obj  reflect.Value
}

func (e OpMissingMapKeyErr) Error() string {
	errMsg := "Expected to find a map key '%s' for path '%s' (%s)"
	return fmt.Sprintf(errMsg, e.Key, e.Path, e.siblingKeysErrStr())
}

func (e OpMissingMapKeyErr) siblingKeysErrStr() string {
	if e.Obj.Len() == 0 {
		return "found no other map keys"
	}

	var keys []string
	for _, key := range e.Obj.MapKeys() {
		if k := dereference(key); k.Kind() == reflect.String {
			keys = append(keys, k.String())
		}
	}

	sort.Sort(sort.StringSlice(keys))

	return "found map keys: '" + strings.Join(keys, "', '") + "'"
}

type OpMissingIndexErr struct {
	Idx  int
	Obj  reflect.Value
	Path Pointer
}

func (e OpMissingIndexErr) Error() string {
	return fmt.Sprintf("Expected to find array index '%d' but found array of length '%d' for path '%s'", e.Idx, e.Obj.Len(), e.Path)
}

type OpMultipleMatchingIndexErr struct {
	Path Pointer
	Idxs []int
}

func (e OpMultipleMatchingIndexErr) Error() string {
	return fmt.Sprintf("Expected to find exactly one matching array item for path '%s' but found %d", e.Path, len(e.Idxs))
}

type OpUnexpectedTokenErr struct {
	Token Token
	Path  Pointer
}

func (e OpUnexpectedTokenErr) Error() string {
	return fmt.Sprintf("Expected to not find token '%T' at path '%s'", e.Token, e.Path)
}
