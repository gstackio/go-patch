package patch_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gstackio/go-patch/patch"
)

var _ = Describe("ArrayIndex", func() {
	dummyPath := MustNewPointerFromString("")

	Describe("Concrete", func() {
		It("returns positive index", func() {
			idx := ArrayIndex{Index: 0, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 1, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 2, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))
		})

		It("wraps around negative index one time", func() {
			idx := ArrayIndex{Index: -0, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: -1, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))

			idx = ArrayIndex{Index: -2, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: -3, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))
		})

		It("does not work with empty arrays", func() {
			idx := ArrayIndex{Index: 0, Modifiers: nil, Array: reflect.ValueOf([]interface{}{}), Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '0' but found array of length '0' for path ''`))

			p := PrevModifier{}
			n := NextModifier{}

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, n}, Array: reflect.ValueOf([]interface{}{}), Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '0' but found array of length '0' for path ''`))
		})

		It("does not work with index out of bounds", func() {
			idx := ArrayIndex{Index: 3, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '3' but found array of length '3' for path ''`))

			idx = ArrayIndex{Index: -4, Modifiers: nil, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '-4' but found array of length '3' for path ''`))
		})

		It("returns previous item when previous modifier is used", func() {
			p := PrevModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p, p, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '-4' but found array of length '3' for path ''`))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p, p, p, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '-5' but found array of length '3' for path ''`))

			idx = ArrayIndex{Index: 2, Modifiers: []Modifier{p, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))
		})

		It("returns next item when next modifier is used", func() {
			n := NextModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{n}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, n}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '3' but found array of length '3' for path ''`))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, n, n}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(MatchError(`Expected to find array index '4' but found array of length '3' for path ''`))
		})

		It("works with multiple previous and next modifiers", func() {
			p := PrevModifier{}
			n := NextModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{p, n}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, n, p}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))
		})

		It("does not support any other modifier except previous and next", func() {
			b := BeforeModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{b}, Array: reflect.ValueOf([]interface{}{1, 2, 3}), Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(MatchError("Expected to find one of the following modifiers: 'prev', 'next', but found modifier 'patch.BeforeModifier'"))
		})
	})
})
