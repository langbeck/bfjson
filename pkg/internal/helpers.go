package internal

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
)

func WritePackage(buf *bytes.Buffer, pkg *types.Package, qf types.Qualifier) {
	if pkg == nil {
		return
	}
	var s string
	if qf != nil {
		s = qf(pkg)
	} else {
		s = pkg.Path()
	}
	if s != "" {
		idx := strings.LastIndexByte(s, '/')
		if idx > 0 {
			s = s[idx+1:]
		}

		buf.WriteString(s)
		buf.WriteByte('.')
	}
}

func WriteType(buf *bytes.Buffer, typ types.Type, qf types.Qualifier) {
	switch t := typ.(type) {
	case *types.Basic:
		buf.WriteString(t.Name())

	case *types.Array:
		fmt.Fprintf(buf, "[%d]", t.Len())
		WriteType(buf, t.Elem(), qf)

	case *types.Slice:
		buf.WriteString("[]")
		WriteType(buf, t.Elem(), qf)

	case *types.Pointer:
		buf.WriteByte('*')
		WriteType(buf, t.Elem(), qf)

	case *types.Map:
		buf.WriteString("map[")
		WriteType(buf, t.Key(), qf)
		buf.WriteByte(']')
		WriteType(buf, t.Elem(), qf)

	case *types.Named:
		obj := t.Obj()
		WritePackage(buf, obj.Pkg(), qf)
		buf.WriteString(obj.Name())

	default:
		// If we got here just implement the missing case
		panic(fmt.Sprintf("unsupported type: %s", t.String()))
	}
}

func TypeString(typ types.Type, qf types.Qualifier) string {
	var buf bytes.Buffer
	WriteType(&buf, typ, qf)
	return buf.String()
}
