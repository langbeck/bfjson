package goparser

import (
	"errors"
	"go/token"

	"golang.org/x/tools/go/packages"
)

var DefaultContext = NewContext()

type Context struct {
	cfg *packages.Config
}

func NewContext() *Context {
	return &Context{
		cfg: &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes,
			Fset: token.NewFileSet(),
		},
	}
}

var (
	ErrNoPackage = errors.New("no packages found")
	ErrTooMany   = errors.New("too many packages found")
)

func (ctx *Context) Import(path string) (*Package, error) {
	pkgs, err := ctx.Load(path)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, ErrNoPackage
	}

	if len(pkgs) > 1 {
		return nil, ErrTooMany
	}

	return pkgs[0], nil
}

func (ctx *Context) Load(path string) ([]*Package, error) {
	ppkgs, err := packages.Load(ctx.cfg, path)
	if err != nil {
		return nil, err
	}

	results := make([]*Package, 0, len(ppkgs))
	for _, ppkg := range ppkgs {
		pkg, err := packageFromTools(ppkg)
		if err != nil {
			return nil, err
		}

		results = append(results, pkg)
	}

	return results, nil
}
