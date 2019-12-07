package support

import "github.com/jinzhu/copier"

// DeepClone deeply clones from 1 interface to another.
func DeepClone(dst, src interface{}) error {
	return copier.Copy(dst, src)
}
