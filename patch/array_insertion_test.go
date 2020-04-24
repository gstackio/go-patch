package patch_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gstackio/go-patch/patch"
)

var _ = Describe("ArrayInsertion", func() {
	Describe("Concrete", func() {
		act := func(insertion ArrayInsertion, array reflect.Value, obj interface{}) (interface{}, error) {
			insertion.Array = array

			idx, err := insertion.Concrete()
			if err != nil {
				return nil, err
			}

			return idx.Update(array, obj), nil
		}

		It("returns specified index when not using any modifiers", func() {
			result, err := act(ArrayInsertion{Index: 1, Modifiers: []Modifier{}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 10, 3}))
		})

		It("returns index adjusted for previous and next modifiers", func() {
			p := PrevModifier{}
			n := NextModifier{}

			result, err := act(ArrayInsertion{Index: 1, Modifiers: []Modifier{p, n, n}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 2, 10}))
		})

		It("returns error if both after and before are used", func() {
			_, err := act(ArrayInsertion{Index: 0, Modifiers: []Modifier{BeforeModifier{}, AfterModifier{}}}, reflect.ValueOf([]interface{}{}), 10)
			Expect(err.Error()).To(Equal("Expected to not find any modifiers after 'before' modifier, but found modifier 'patch.AfterModifier'"))

			_, err = act(ArrayInsertion{Index: 0, Modifiers: []Modifier{AfterModifier{}, BeforeModifier{}}}, reflect.ValueOf([]interface{}{}), 10)
			Expect(err.Error()).To(Equal("Expected to not find any modifiers after 'after' modifier, but found modifier 'patch.BeforeModifier'"))

			_, err = act(ArrayInsertion{Index: 0, Modifiers: []Modifier{AfterModifier{}, PrevModifier{}}}, reflect.ValueOf([]interface{}{}), 10)
			Expect(err.Error()).To(Equal("Expected to not find any modifiers after 'after' modifier, but found modifier 'patch.PrevModifier'"))
		})

		It("returns (0, true) when inserting in the beginning", func() {
			result, err := act(ArrayInsertion{Index: 0, Modifiers: []Modifier{BeforeModifier{}}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{10, 1, 2, 3}))

			result, err = act(ArrayInsertion{Index: 0, Modifiers: []Modifier{AfterModifier{}}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 10, 2, 3}))
		})

		It("returns (last, true) when inserting in the end", func() {
			result, err := act(ArrayInsertion{Index: 2, Modifiers: []Modifier{AfterModifier{}}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 2, 3, 10}))

			result, err = act(ArrayInsertion{Index: -1, Modifiers: []Modifier{AfterModifier{}}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 2, 3, 10}))
		})

		It("returns (mid, true) when inserting in the middle", func() {
			result, err := act(ArrayInsertion{Index: 1, Modifiers: []Modifier{AfterModifier{}}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 2, 10, 3}))

			result, err = act(ArrayInsertion{Index: 1, Modifiers: []Modifier{BeforeModifier{}}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 10, 2, 3}))

			result, err = act(ArrayInsertion{Index: 2, Modifiers: []Modifier{BeforeModifier{}}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 2, 10, 3}))
		})

		It("returns index adjusted for previous, next modifiers and before modifier", func() {
			p := PrevModifier{}
			n := NextModifier{}
			b := BeforeModifier{}

			result, err := act(ArrayInsertion{Index: 1, Modifiers: []Modifier{p, n, n, b}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 2, 10, 3}))
		})

		It("returns index adjusted for previous, next modifiers and after modifier", func() {
			p := PrevModifier{}
			n := NextModifier{}
			a := AfterModifier{}

			result, err := act(ArrayInsertion{Index: 1, Modifiers: []Modifier{p, n, n, a}}, reflect.ValueOf([]interface{}{1, 2, 3}), 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]interface{}{1, 2, 3, 10}))
		})
	})
})
