package goparser

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

var ErrPackageMain = errors.New(`main packages are not supported`)

var packageMain = "main"

type Package struct {
	tpkg  *types.Package
	tinfo *types.Info

	structs       []*Struct
	objects       []Object
	objectForName map[string]Object
	objectForType map[types.Type]Object
	ctx           *Context
}

func packageFromTools(ppkg *packages.Package) (*Package, error) {
	if len(ppkg.Errors) > 0 {
		return nil, ppkg.Errors[0]
	}

	if ppkg.Name == packageMain {
		return nil, ErrPackageMain
	}

	pkg := &Package{
		tinfo: ppkg.TypesInfo,
		tpkg:  ppkg.Types,

		objectForType: make(map[types.Type]Object),
		objectForName: make(map[string]Object),
		objects:       make([]Object, 0, 50),
		structs:       make([]*Struct, 0, 50),
	}

	// Look for type declarations in any *ast.File
	for _, f := range ppkg.Syntax {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if gd.Tok != token.TYPE {
				continue
			}

			// Parse annotations for the whole type group
			groupAnnotations := parseAnnotations(nil, gd.Doc)
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// Resolve the type.Object associated with *ast.Ident using the
				// information populated by types.Config.Check
				obj := ppkg.TypesInfo.ObjectOf(ts.Name)
				if obj == nil || !obj.Exported() {
					continue
				}

				named, ok := obj.Type().(*types.Named)
				if !ok {
					continue
				}

				// Parse type annotations and merge it with the group ones
				annotations := parseAnnotations(groupAnnotations, ts.Doc)

				switch named.Underlying().(type) {
				case *types.Struct:
					s := NewStruct(pkg, ts.Name, obj, annotations)
					pkg.structs = append(pkg.structs, s)
					pkg.objects = append(pkg.objects, s)
					pkg.objectForName[s.Name()] = s
					pkg.objectForType[s.Type()] = s

				default:
					o := NewObject(pkg, ts.Name, obj, annotations)
					pkg.objects = append(pkg.objects, o)
					pkg.objectForName[o.Name()] = o
					pkg.objectForType[o.Type()] = o
				}
			}
		}
	}

	return pkg, nil
}

func (pkg *Package) Context() *Context {
	return pkg.ctx
}

func (pkg *Package) ObjectForName(name string) Object {
	// log.Printf("ObjectForName: %s", name)
	return pkg.objectForName[name]
}

func (pkg *Package) ObjectForType(typ types.Type) Object {
	// log.Printf("ObjectForType: %-20T\t%v", typ, typ.String())
	return pkg.objectForType[typ]
}

func (pkg *Package) Path() string {
	return pkg.tpkg.Path()
}

func (pkg *Package) Structs() []*Struct {
	return pkg.structs
}
