package support

import (
	"testing"

	"github.com/appist/appy/test"
)

func TestDeepClone(t *testing.T) {
	assert := test.NewAssert(t)

	type User struct {
		Email string
		Name  string
	}

	type Employee struct {
		Email string
		Name  string
		Role  string
	}

	user := User{Email: "john_doe@gmail.com", Name: "John Doe"}
	employee := Employee{}
	DeepClone(&employee, &user)
	assert.Equal("john_doe@gmail.com", employee.Email)
	assert.Equal("John Doe", employee.Name)

	employees := []Employee{}
	DeepClone(&employees, &user)
	assert.Equal(1, len(employees))
	assert.Equal("john_doe@gmail.com", employees[0].Email)
	assert.Equal("John Doe", employees[0].Name)

	users := []User{
		{Email: "john_doe1@gmail.com", Name: "John Doe 1"},
		{Email: "john_doe2@gmail.com", Name: "John Doe 2"},
	}
	employees = []Employee{}
	DeepClone(&employees, &users)
	assert.Equal(2, len(employees))
	assert.Equal("john_doe1@gmail.com", employees[0].Email)
	assert.Equal("John Doe 1", employees[0].Name)
	assert.Equal("john_doe2@gmail.com", employees[1].Email)
	assert.Equal("John Doe 2", employees[1].Name)
}

func TestArrayContains(t *testing.T) {
	assert := test.NewAssert(t)
	tests := []struct {
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

	for _, tt := range tests {
		assert.Equal(tt.expected, ArrayContains(tt.arr, tt.val))
	}
}
