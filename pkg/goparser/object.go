package goparser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"

	"github.com/langbeck/bfjson/pkg/internal"
)

type StructField struct {
	Name     string
	Embedded bool
	Type     types.Type
	Tag      string
}

type Struct struct {
	*objectBase
	fields []StructField

	named      *types.Named
	underlying *types.Struct
}

func NewStruct(pkg *Package, ident *ast.Ident, obj types.Object, flags Annotations) *Struct {
	o := newObjectBase("NewStruct", pkg, ident, obj, flags)

	named, _ := obj.Type().(*types.Named)
	if named == nil {
		panic("NewStruct: obj.Type() must be *types.Named")
	}

	ts, _ := named.Underlying().(*types.Struct)
	if ts == nil {
		panic("NewStruct: obj.Type().Underlying() must be *types.Struct")
	}

	fields := make([]StructField, 0, ts.NumFields())
	for nf := 0; nf < ts.NumFields(); nf++ {
		field := ts.Field(nf)
		fields = append(fields, StructField{
			Embedded: field.Embedded(),
			Name:     field.Name(),
			Type:     field.Type(),
			Tag:      ts.Tag(nf),
		})
	}

	return &Struct{
		objectBase: o,
		named:      named,
		underlying: ts,
		fields:     fields,
	}
}

func (s *Struct) Fields() []StructField {
	return append([]StructField{}, s.fields...)
}

func (s *Struct) Underlying() *types.Struct {
	return s.underlying
}

type Object interface {
	Implements(i *types.Interface) bool
	HasAnnotation(string) bool

	Package() *Package
	Type() types.Type

	Name() string
	QualifiedName(qf types.Qualifier) string
}

func NewObject(pkg *Package, ident *ast.Ident, obj types.Object, flags Annotations) Object {
	return newObjectBase("NewObject", pkg, ident, obj, flags)
}

type objectBase struct {
	ident *ast.Ident
	tobj  types.Object
	flags Annotations
	pkg   *Package
}

func newObjectBase(ctxname string, pkg *Package, ident *ast.Ident, obj types.Object, flags Annotations) *objectBase {
	switch {
	case pkg == nil:
		panic(fmt.Sprintf("%s: nil *Package", ctxname))

	case ident == nil:
		panic(fmt.Sprintf("%s: nil *ast.Ident", ctxname))

	case obj == nil:
		panic(fmt.Sprintf("%s: nil type.Object", ctxname))
	}

	return &objectBase{
		pkg:   pkg,
		ident: ident,
		tobj:  obj,
		flags: flags,
	}
}

func (o *objectBase) Type() types.Type {
	return o.tobj.Type()
}

func (o *objectBase) HasAnnotation(s string) bool {
	for _, a := range o.flags {
		if s == a {
			return true
		}
	}

	return false
}

func (o *objectBase) Implements(i *types.Interface) (b bool) {
	if o == nil {
		return false
	}

	if types.Implements(o.tobj.Type(), i) {
		return true
	}

	return types.Implements(types.NewPointer(o.tobj.Type()), i)
}

func (o *objectBase) Package() *Package {
	return o.pkg
}

func (o *objectBase) Name() string {
	return o.tobj.Name()
}

func (o *objectBase) QualifiedName(qf types.Qualifier) string {
	var buf bytes.Buffer
	internal.WritePackage(&buf, o.tobj.Pkg(), qf)
	buf.WriteString(o.tobj.Name())
	return buf.String()
}
