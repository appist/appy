package support

import (
	"reflect"

	"github.com/jinzhu/copier"
)

// Copy deeply clones from 1 interface to another.
func Copy(dst, src interface{}) error {
	return copier.Copy(dst, src)
}

// Contains is a helper function to check if a value is in a slice.
func Contains(arr interface{}, val interface{}) bool {
	arrT := reflect.TypeOf(arr)
	valT := reflect.TypeOf(val)
	if (arrT.Kind().String() != "array" && arrT.Kind().String() != "slice") ||
		arrT.Elem().String() != valT.Kind().String() {
		return false
	}

	switch arr := arr.(type) {
	case []bool:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []byte:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []complex64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []complex128:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []float32:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []float64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int8:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int16:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int32:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint16:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint32:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uintptr:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []string:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	}

	return false
}
