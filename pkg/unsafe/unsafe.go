package unsafe

import (
	"reflect"
	"unsafe"
)

// Just a type alias to the actual unsafe.Pointer
type Pointer = unsafe.Pointer

func BytesToString(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	stringHeader := reflect.StringHeader{Data: sliceHeader.Data, Len: sliceHeader.Len}
	return *(*string)(unsafe.Pointer(&stringHeader))
}
