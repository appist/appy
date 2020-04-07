package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type ArraySuite struct {
	test.Suite
}

func (s *ArraySuite) TestArrayContains() {
	tt := []struct {
		arr      interface{}
		val      interface{}
		expected bool
	}{
		{[]bool{}, true, false},
		{[]bool{true, true}, false, false},
		{[]bool{true, true}, true, true},
		{[]byte{}, 80, false},
		{[]byte{80, 81}, byte(82), false},
		{[]byte{80, 81}, byte(81), true},
		{[]complex64{}, 1 + 2i, false},
		{[]complex64{1 + 2i, 2 + 3i}, complex64(3 + 4i), false},
		{[]complex64{1 + 2i, 2 + 3i}, complex64(2 + 3i), true},
		{[]complex128{}, 1 + 2i, false},
		{[]complex128{1 + 2i, 2 + 3i}, complex128(3 + 4i), false},
		{[]complex128{1 + 2i, 2 + 3i}, complex128(2 + 3i), true},
		{[]float32{}, 0.1, false},
		{[]float32{0.1, 0.2}, float32(0.3), false},
		{[]float32{0.1, 0.2}, float32(0.1), true},
		{[]float64{}, 0.1, false},
		{[]float64{0.1, 0.2}, float64(0.3), false},
		{[]float64{0.1, 0.2}, float64(0.1), true},
		{[]int{}, 1, false},
		{[]int{1, 2}, int(3), false},
		{[]int{1, 2}, int(1), true},
		{[]int8{}, 1, false},
		{[]int8{1, 2}, int8(3), false},
		{[]int8{1, 2}, int8(1), true},
		{[]int16{}, 1, false},
		{[]int16{1, 2}, int16(3), false},
		{[]int16{1, 2}, int16(1), true},
		{[]int32{}, 1, false},
		{[]int32{1, 2}, int32(3), false},
		{[]int32{1, 2}, int32(1), true},
		{[]int64{}, 1, false},
		{[]int64{1, 2}, int64(3), false},
		{[]int64{1, 2}, int64(1), true},
		{[]uint{}, 1, false},
		{[]uint{1, 2}, uint(3), false},
		{[]uint{1, 2}, uint(1), true},
		{[]uint16{}, 1, false},
		{[]uint16{1, 2}, uint16(3), false},
		{[]uint16{1, 2}, uint16(1), true},
		{[]uint32{}, 1, false},
		{[]uint32{1, 2}, uint32(3), false},
		{[]uint32{1, 2}, uint32(1), true},
		{[]uint64{}, 1, false},
		{[]uint64{1, 2}, uint64(3), false},
		{[]uint64{1, 2}, uint64(1), true},
		{[]uintptr{}, 1, false},
		{[]uintptr{1, 2}, uintptr(3), false},
		{[]uintptr{1, 2}, uintptr(1), true},
		{[]string{}, "a", false},
		{[]string{"a", "b"}, "c", false},
		{[]string{"a", "b"}, "a", true},
	}

	for _, t := range tt {
		s.Equal(t.expected, ArrayContains(t.arr, t.val))
	}
}

func TestArraySuite(t *testing.T) {
	test.Run(t, new(ArraySuite))
}
