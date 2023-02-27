package basics

import (
	"github.com/langbeck/bfjson/pkg/unsafe"
	"github.com/valyala/fastjson"
)

func DecodeString(v *fastjson.Value, dst *string) error {
	if v.Type() == fastjson.TypeNull {
		return nil
	}

	sb, err := v.StringBytes()
	if err != nil {
		return err
	}

	*dst = unsafe.BytesToString(sb)
	return nil
}

func DecodeSliceOfString(v *fastjson.Value, dst *[]string) error {
	if v.Type() == fastjson.TypeNull {
		*dst = nil
		return nil
	}

	arr, err := v.Array()
	if err != nil {
		return err
	}

	slice := make([]string, len(arr))
	for idx, item := range arr {
		sb, err := item.StringBytes()
		if err != nil {
			return err
		}

		slice[idx] = unsafe.BytesToString(sb)
	}

	*dst = slice
	return nil
}

func DecodeInt(v *fastjson.Value, dst *int) error {
	if v.Type() == fastjson.TypeNull {
		return nil
	}

	n, err := v.Int()
	if err != nil {
		return err
	}

	*dst = n
	return nil
}

func DecodePtrInt(v *fastjson.Value, dst **int) error {
	if v.Type() == fastjson.TypeNull {
		*dst = nil
		return nil
	}

	n, err := v.Int()
	if err != nil {
		return err
	}

	*dst = &n
	return nil
}

func DecodeSliceOfInt(v *fastjson.Value, dst *[]int) error {
	if v.Type() == fastjson.TypeNull {
		*dst = nil
		return nil
	}

	arr, err := v.Array()
	if err != nil {
		return err
	}

	slice := make([]int, len(arr))
	for idx, item := range arr {
		n, err := item.Int()
		if err != nil {
			return err
		}

		slice[idx] = n
	}

	*dst = slice
	return nil
}

func DecodeFloat64(v *fastjson.Value, dst *float64) error {
	if v.Type() == fastjson.TypeNull {
		return nil
	}

	f, err := v.Float64()
	if err != nil {
		return err
	}

	*dst = f
	return nil
}
